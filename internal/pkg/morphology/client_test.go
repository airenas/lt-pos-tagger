package morphology

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
	var resp api.TaggerResult
	rb, _ := json.Marshal(resp)
	cl, server := initServer(t, "/", string(rb), 200)
	defer server.Close()

	r, err := cl.Process("olia", &api.SegmenterResult{Seg: [][]int{{1}}, S: [][]int{{1}}})

	assert.Nil(t, err)
	assert.NotNil(t, r)
}

func TestProcess_WrongCode_Fails(t *testing.T) {
	cl, server := initServer(t, "/", "", 400)
	defer server.Close()

	r, err := cl.Process("olia", &api.SegmenterResult{Seg: [][]int{{1}}, S: [][]int{{1}}})
	assert.NotNil(t, err)
	assert.Nil(t, r)
}

func TestProcess_Retry(t *testing.T) {
	cl, server := initServer(t, "/", "", 429)
	defer server.Close()
	cl.timeOut = 500 * time.Millisecond
	r, err := cl.Process("olia", &api.SegmenterResult{Seg: [][]int{{1}}, S: [][]int{{1}}})
	assert.NotNil(t, err)
	assert.Nil(t, r)
}

func TestProcess_NoTex_Fails(t *testing.T) {
	rb, _ := json.Marshal(api.TaggerResult{})
	cl, server := initServer(t, "/", string(rb), 200)
	defer server.Close()

	r, err := cl.Process("", &api.SegmenterResult{})
	assert.NotNil(t, err)
	assert.Nil(t, r)
}

func TestProcess_WrongLex_Fails(t *testing.T) {
	rb, _ := json.Marshal(api.TaggerResult{})
	cl, server := initServer(t, "/", string(rb), 200)
	defer server.Close()

	r, err := cl.Process("olia", nil)
	assert.NotNil(t, err)
	assert.Nil(t, r)

	r, err = cl.Process("olia", &api.SegmenterResult{})
	assert.NotNil(t, err)
	assert.Nil(t, r)
}
