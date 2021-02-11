package morphology

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
	api := Client{}
	api.httpclient = server.Client()
	api.url = server.URL
	return &api, server
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

func TestValidateResp(t *testing.T) {
	tResp := httptest.NewRecorder()
	tResp.Body = bytes.NewBuffer([]byte("olia"))
	tResp.Code = 200
	err := ValidateResponse(tResp.Result())
	assert.Nil(t, err)
}

func TestValidateResp_Fail(t *testing.T) {
	tResp := httptest.NewRecorder()
	tResp.Body = bytes.NewBuffer([]byte("err olia"))
	tResp.Code = 400
	err := ValidateResponse(tResp.Result())
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "err olia")
}

func TestValidateResp_FailLong(t *testing.T) {
	tResp := httptest.NewRecorder()
	ls := ""
	for len(ls) < 100 {
		ls = ls + "abc"
	}
	tResp.Body = bytes.NewBuffer([]byte(ls))
	tResp.Code = 400
	err := ValidateResponse(tResp.Result())
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "abc")
	assert.Contains(t, err.Error(), "...")
}
