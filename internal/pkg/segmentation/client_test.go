package segmentation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
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
	api.lexLock = &sync.Mutex{}
	return &api, server
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
