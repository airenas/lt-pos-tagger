package segmentation

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/lt-pos-tagger/internal/pkg/api"
	"github.com/airenas/lt-pos-tagger/internal/pkg/morphology"
	"github.com/airenas/lt-pos-tagger/internal/pkg/utils"
	"github.com/pkg/errors"
	"mvdan.cc/xurls/v2"
)

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
		return nil, errors.New("No lex URL")
	}
	res.url = url
	res.httpclient = http.DefaultClient
	res.rateLimit = make(chan struct{}, 1)

	return &res, nil
}

//Process invokes ws
func (t *Client) Process(data string) (*api.SegmenterResult, error) {
	if len([]rune(data)) == 1 {
		return &api.SegmenterResult{Seg: [][]int{{0, 1}}, P: [][]int{{0, 1}}, S: [][]int{{0, 1}}}, nil
	}

	// lex fails if several requests go simultaneously
	select {
	case t.rateLimit <- struct{}{}:
	case <-time.After(20 * time.Second):
		return nil, utils.ErrTooBusy
	}
	defer func() { <-t.rateLimit }()

	ctx, cancelF := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelF()
	bytesData := []byte(data)
	req, err := http.NewRequest(http.MethodPost, t.url, bytes.NewBuffer(bytesData))
	if err != nil {
		return nil, errors.Wrapf(err, "can't prepare request to '%s'", t.url)
	}
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)

	resp, err := t.httpclient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "can't invoke lex %s", t.url)
	}
	err = morphology.ValidateResponse(resp)
	if err != nil {
		return nil, errors.Wrap(err, "can't invoke lex")
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	var res api.SegmenterResult
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, errors.Wrap(err, "can't decode response")
	}
	goapp.Log.Debugf("Lex: %v", res.Seg)
	res.Seg = fixSegments(res.Seg, data)
	return &res, nil
}

var (
	fixSymbolsMap map[rune]bool
	urlRegexp     *regexp.Regexp
)

func init() {
	fixSymbolsMap = make(map[rune]bool)
	for _, r := range "-‘\"–‑/:;`−≤≥⁰" {
		fixSymbolsMap[r] = true
	}
	urlRegexp = xurls.Relaxed()
}

func fixSegments(seg [][]int, data string) [][]int {
	res := make([][]int, 0)
	sr := []rune(data)
	for _, s := range seg {
		if s[1] == 1 {
			res = append(res, s)
			continue
		}
		rw := sr[s[0] : s[0]+s[1]]
		st := string(rw)
		if isURL(st) || utils.IsNumber(st) {
			res = append(res, s)
			continue
		}

		f := 0
		for i, r := range rw {
			if fixSymbolsMap[r] {
				if f != i {
					res = append(res, []int{s[0] + f, i - f})
				}
				res = append(res, []int{s[0] + i, 1})
				f = i + 1
			}
		}
		if f < len(rw) {
			res = append(res, []int{s[0] + f, len(rw) - f})
		}
	}
	return res
}

func isURL(s string) bool {
	return urlRegexp.FindString(s) == s
}
