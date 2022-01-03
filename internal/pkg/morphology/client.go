package morphology

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/lt-pos-tagger/internal/pkg/api"
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
}

//NewClient creates a tagger client
func NewClient(url string) (*Client, error) {
	res := Client{}
	if url == "" {
		return nil, errors.New("No morphology URL")
	}
	res.url = url
	res.httpclient = http.DefaultClient
	res.rateLimit = make(chan struct{}, 10)
	return &res, nil
}

//Process invokes ws
func (t *Client) Process(text string, data *api.SegmenterResult) (*api.TaggerResult, error) {
	// allow only 10 paraller requests to morph as it fails to process more
	select {
	case t.rateLimit <- struct{}{}:
	case <-time.After(20 * time.Second):
		return nil, errors.Errorf("morphology too busy, timeouted")
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
	ctx, cancelF := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelF()
	req, err := http.NewRequest(http.MethodPost, t.url, bytes.NewBuffer(bytesData))
	if err != nil {
		return nil, errors.Wrapf(err, "can't prepare request to '%s'", t.url)
	}
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)
	//goapp.Log.Debugf("Input: %s", string(bytesData))
	resp, err := t.httpclient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "can't invoke tagger %s", t.url)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	err = ValidateResponse(resp)
	if err != nil {
		return nil, errors.Wrap(err, "can't invoke tagger")
	}
	var result api.TaggerResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, errors.Wrap(err, "can't decode response")
	}
	return &result, nil
}

//ValidateResponse returns error if code is not in [200, 299]
func ValidateResponse(resp *http.Response) error {
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		bodyBytes := make([]byte, 101)
		n, err := resp.Body.Read(bodyBytes)
		if (err != nil) {
			goapp.Log.Error(err)
		}
		trimS := ""
		if n > 100 {
			n = 100
			trimS = "..."
		}
		msg := fmt.Sprintf("wrong response code from server. Code: %d\n%s",
			resp.StatusCode, string(bodyBytes[:n])+trimS)
		return errors.New(msg)
	}
	return nil
}
