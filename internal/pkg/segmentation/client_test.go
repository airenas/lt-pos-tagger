package segmentation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/airenas/lt-pos-tagger/internal/pkg/api"
	"github.com/stretchr/testify/assert"
)

func initServer(t *testing.T, urlStr, resp string, code int) (*Client, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.URL.String(), urlStr)
		rw.WriteHeader(code)
		rw.Write([]byte(resp))
	}))
	// Use Client & URL from our local test server
	api, _ := NewClient(server.URL)
	api.httpclient = server.Client()
	api.url = server.URL
	api.rateLimit = make(chan struct{}, 1)
	return api, server
}

func TestNew(t *testing.T) {
	c, err := NewClient("url.url")
	assert.Nil(t, err)
	assert.NotNil(t, c)
}

func TestNew_Fail(t *testing.T) {
	c, err := NewClient("")
	assert.NotNil(t, err)
	assert.Nil(t, c)
}

func TestProcess(t *testing.T) {
	var resp api.SegmenterResult
	rb, _ := json.Marshal(resp)
	cl, server := initServer(t, "/", string(rb), 200)
	defer server.Close()

	r, err := cl.Process("olia")

	assert.Nil(t, err)
	assert.NotNil(t, r)
}

func TestProcess_WrongCode_Fails(t *testing.T) {
	cl, server := initServer(t, "/", "", 400)
	defer server.Close()

	r, err := cl.Process("olia")
	assert.NotNil(t, err)
	assert.Nil(t, r)
}

func TestProcess_Retry(t *testing.T) {
	cl, server := initServer(t, "/", "", 429)
	defer server.Close()
	cl.timeOut = 500 * time.Millisecond
	r, err := cl.Process("olia")
	assert.NotNil(t, err)
	assert.Nil(t, r)
}

func TestProcess_OneLetter(t *testing.T) {

	cl, server := initServer(t, "/", "a", 200)
	defer server.Close()

	r, err := cl.Process("a")
	assert.Nil(t, err)
	if assert.NotNil(t, r) {
		assert.Equal(t, [][]int{{0, 1}}, r.Seg)
		assert.Equal(t, [][]int{{0, 1}}, r.P)
		assert.Equal(t, [][]int{{0, 1}}, r.S)
	}
}

func TestFixSegments(t *testing.T) {
	tests := []struct {
		v [][]int
		e [][]int
		s string
		i string
	}{
		{v: [][]int{{0, 2}, {3, 2}}, s: "aa bb", e: [][]int{{0, 2}, {3, 2}}, i: "simple"},
		{v: [][]int{{0, 5}}, s: "aa-bb", e: [][]int{{0, 2}, {2, 1}, {3, 2}}, i: "splits '-'"},
		{v: [][]int{{0, 5}}, s: "aa–bb", e: [][]int{{0, 2}, {2, 1}, {3, 2}}, i: "splits '–'"},
		{v: [][]int{{0, 5}}, s: "aa≤bb", e: [][]int{{0, 2}, {2, 1}, {3, 2}}, i: "splits '≤'"},
		{v: [][]int{{0, 5}}, s: "aa≥bb", e: [][]int{{0, 2}, {2, 1}, {3, 2}}, i: "splits '≥'"},
		{v: [][]int{{0, 5}}, s: "aa⁰bb", e: [][]int{{0, 2}, {2, 1}, {3, 2}}, i: "splits '⁰'"},
		{v: [][]int{{0, 2}, {3, 1}, {5, 2}}, s: "aa - bb", e: [][]int{{0, 2}, {3, 1}, {5, 2}}, i: "leaves '-'"},
		{v: [][]int{{0, 5}}, s: "aa/bb", e: [][]int{{0, 2}, {2, 1}, {3, 2}}, i: "splits '/'"},
		{v: [][]int{{0, 3}}, s: "\"bb", e: [][]int{{0, 1}, {1, 2}}, i: "splits '\"'"},
		{v: [][]int{{0, 16}}, s: "https://delfi.lt", e: [][]int{{0, 16}}, i: "leaves url"},
		{v: [][]int{{0, 17}}, s: "delfi.lt/20/20/20", e: [][]int{{0, 17}}, i: "leaves url"},
		{v: [][]int{{0, 3}}, s: "-10", e: [][]int{{0, 3}}, i: "leaves int number"},
		{v: [][]int{{0, 3}}, s: "+10", e: [][]int{{0, 3}}, i: "leaves int number"},
		{v: [][]int{{0, 2}}, s: "10", e: [][]int{{0, 2}}, i: "leaves int number"},
		{v: [][]int{{0, 6}}, s: "-10,15", e: [][]int{{0, 6}}, i: "leaves float number"},
		{v: [][]int{{0, 6}}, s: "-10.15", e: [][]int{{0, 6}}, i: "leaves float number"},
		{v: [][]int{{0, 12}}, s: "-1.12312e+15", e: [][]int{{0, 12}}, i: "leaves scientific format"},
		{v: [][]int{{0, 11}}, s: "1.12312e+15", e: [][]int{{0, 11}}, i: "leaves scientific format"},
		{v: [][]int{{0, 3}}, s: "a:2", e: [][]int{{0, 1}, {1, 1}, {2, 1}}, i: "parses ':'"},
		{v: [][]int{{0, 4}}, s: "10;2", e: [][]int{{0, 2}, {2, 1}, {3, 1}}, i: "parses ';'"},
		{v: [][]int{{0, 7}}, s: "'aa'bb'", e: [][]int{{0, 1}, {1, 2}, {3, 1}, {4, 2}, {6, 1}}, i: "splits '\\''"},
		{v: [][]int{{0, 1}, {1, 3}}, s: "'Aa'", e: [][]int{{0, 1}, {1, 2}, {3, 1}}, i: "splits '\\''"},
	}
	for _, tt := range tests {
		t.Run(tt.i, func(t *testing.T) {
			assert.Equal(t, tt.e, fixSegments(tt.v, tt.s), "Fail %s", tt.i)
		})
	}
}
