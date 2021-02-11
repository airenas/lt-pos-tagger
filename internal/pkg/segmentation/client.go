package segmentation

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/airenas/lt-pos-tagger/internal/pkg/api"
	"github.com/airenas/lt-pos-tagger/internal/pkg/morphology"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"mvdan.cc/xurls/v2"
)

//Client comunicates with tagger server
type Client struct {
	httpclient *http.Client
	url        string
	lexLock    *sync.Mutex
}

//NewClient creates a tagger client
func NewClient(url string) (*Client, error) {
	res := Client{}
	if url == "" {
		return nil, errors.New("No lex URL")
	}
	res.url = url
	res.httpclient = http.DefaultClient
	res.lexLock = &sync.Mutex{}

	return &res, nil
}

//Process invokes ws
func (t *Client) Process(data string) (*api.SegmenterResult, error) {
	if len([]rune(data)) == 1 {
		return &api.SegmenterResult{Seg: [][]int{{0, 1}}, P: [][]int{{0, 1}}, S: [][]int{{0, 1}}}, nil
	}

	t.lexLock.Lock()
	defer t.lexLock.Unlock()

	logrus.Debug("Process lex")
	bytesData := []byte(data)
	resp, err := t.httpclient.Post(t.url, "application/json", bytes.NewBuffer(bytesData))
	if err != nil {
		return nil, errors.Wrapf(err, "Can't invoke lex %s", t.url)
	}
	err = morphology.ValidateResponse(resp)
	if err != nil {
		return nil, errors.Wrap(err, "Can't invoke lex")
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	var result api.SegmenterResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, errors.Wrap(err, "Can't decode response")
	}
	result.Seg = fixSegments(result.Seg, data)
	return &result, nil
}

var fixSymbolsMap map[rune]bool

func init() {
	fixSymbolsMap = make(map[rune]bool)
	for _, r := range "-‘\"–‑/" {
		fixSymbolsMap[r] = true
	}
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
		if isURL(string(rw)) {
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
	rxRelaxed := xurls.Relaxed()
	return rxRelaxed.FindString(s) == s
}
