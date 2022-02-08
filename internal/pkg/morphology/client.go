package morphology

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/lt-pos-tagger/internal/pkg/api"
	"github.com/airenas/lt-pos-tagger/internal/pkg/utils"
	"github.com/pkg/errors"
)

type requestAnotations struct {
	Lex *api.SegmenterResult `json:"lex"`
}

type request struct {
	Scope       string            `json:"scope"`
	Body        string            `json:"body"`
	Annotations requestAnotations `json:"annotations"`
}

//Client comunicates with tagger server
type Client struct {
	httpclient *http.Client
	url        string
	rateLimit  chan struct{}
	timeOut    time.Duration
}

//NewClient creates a tagger client
func NewClient(url string) (*Client, error) {
	res := Client{}
	if url == "" {
		return nil, errors.New("No morphology URL")
	}
	res.url = url
	res.httpclient = &http.Client{Transport: newTransport()}
	res.rateLimit = make(chan struct{}, 10)
	res.timeOut = time.Second * 20
	return &res, nil
}

func newTransport() http.RoundTripper {
	res := http.DefaultTransport.(*http.Transport).Clone()
	res.MaxIdleConns = 20
	res.MaxConnsPerHost = 20
	res.MaxIdleConnsPerHost = 20
	return res
}

//Process invokes ws
func (t *Client) Process(text string, data *api.SegmenterResult) (*api.TaggerResult, error) {
	// allow only 10 paraller requests to morph as it fails to process more
	select {
	case t.rateLimit <- struct{}{}:
	case <-time.After(t.timeOut):
		return nil, utils.ErrTooBusy
	}
	defer func() { <-t.rateLimit }()

	goapp.Log.Debug("Process tagger")
	if text == "" {
		return nil, errors.Errorf("no text")
	}
	if data == nil {
		return nil, errors.Errorf("no lex data")
	}
	if len(data.Seg) == 0 || len(data.S) == 0 {
		return nil, errors.Errorf("wrong lex data")
	}
	reqData := request{Scope: "all", Body: text, Annotations: requestAnotations{Lex: data}}
	bytesData, err := json.Marshal(reqData)
	if err != nil {
		return nil, errors.Wrap(err, "can't marshal data")
	}
	ctx, cancelF := context.WithTimeout(context.Background(), t.timeOut)
	defer cancelF()

	var result api.TaggerResult

	oneCall := func(ctx context.Context, result *api.TaggerResult) (bool, error) {
		req, err := http.NewRequest(http.MethodPost, t.url, bytes.NewBuffer(bytesData))
		if err != nil {
			return false, errors.Wrapf(err, "can't prepare request to '%s'", t.url)
		}
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctx)
		//goapp.Log.Debugf("Input: %s", string(bytesData))
		resp, err := t.httpclient.Do(req)
		if err != nil {
			return true, errors.Wrapf(err, "can't invoke tagger %s", t.url)
		}
		defer func() {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}()

		err = goapp.ValidateHTTPResp(resp, 100)
		if err != nil {
			return utils.IsRetryCode(resp.StatusCode), errors.Wrap(err, "can't invoke tagger")
		}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return true, errors.Wrap(err, "can't decode response")
		}
		return false, nil
	}

	for _, st := range utils.ExpBackoffList {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-utils.RandomWait(st):
		}
		var retry bool
		retry, err = oneCall(ctx, &result)
		if !retry {
			if err != nil {
				return nil, err
			}
			break
		}
		if err != nil {
			goapp.Log.Warn(err)
		}
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}
