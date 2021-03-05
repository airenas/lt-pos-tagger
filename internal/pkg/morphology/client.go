package morphology

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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
	rateLimit  chan bool
}

//NewClient creates a tagger client
func NewClient(url string) (*Client, error) {
	res := Client{}
	if url == "" {
		return nil, errors.New("No morphology URL")
	}
	res.url = url
	res.httpclient = http.DefaultClient
	res.rateLimit = make(chan bool, 10)
	return &res, nil
}

//Process invokes ws
func (t *Client) Process(text string, data *api.SegmenterResult) (*api.TaggerResult, error) {
	t.rateLimit <- true
	defer func() { <-t.rateLimit }()

	goapp.Log.Debug("Process tagger")
	if text == "" {
		return nil, errors.Errorf("No text")
	}
	if data == nil {
		return nil, errors.Errorf("No lex data")
	}
	if len(data.Seg) == 0 || len(data.S) == 0 {
		return nil, errors.Errorf("Wrong lex data")
	}
	req := request{Scope: "all", Body: text, Annotations: requestAnotations{Lex: data}}
	bytesData, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "Can't marshal data")
	}
	//goapp.Log.Debugf("Input: %s", string(bytesData))
	resp, err := t.httpclient.Post(t.url, "application/json", bytes.NewBuffer(bytesData))
	if err != nil {
		return nil, errors.Wrapf(err, "Can't invoke tagger %s", t.url)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	err = ValidateResponse(resp)
	if err != nil {
		return nil, errors.Wrap(err, "Can't invoke tagger")
	}
	var result api.TaggerResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, errors.Wrap(err, "Can't decode response")
	}
	return &result, nil
}

//ValidateResponse returns error if code is not in [200, 299]
func ValidateResponse(resp *http.Response) error {
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		trimS := ""
		if len(bodyBytes) > 100 {
			bodyBytes = bodyBytes[:100]
			trimS = "..."
		}
		msg := fmt.Sprintf("Wrong response code from server. Code: %d\n%s",
			resp.StatusCode, string(bodyBytes)+trimS)
		return errors.New(msg)
	}
	return nil
}
