//go:build integration
// +build integration

package integration

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/airenas/lt-pos-tagger/internal/pkg/service"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type config struct {
	url          string
	semantikaURL string
	httpclient   *http.Client
}

var cfg config

func TestMain(m *testing.M) {
	cfg.url = getEnvOrFail("TAGGER_URL")
	cfg.semantikaURL = getEnvOrFail("MORPHOLOGY_URL")
	cfg.httpclient = &http.Client{Timeout: time.Second}

	tCtx, cf := context.WithTimeout(context.Background(), time.Second*20)
	defer cf()
	waitForOpen(tCtx, cfg.url)
	waitForOpen(tCtx, cfg.semantikaURL)

	os.Exit(m.Run())
}

func waitForOpen(ctx context.Context, URL string) {
	u, err := url.Parse(URL)
	if err != nil {
		log.Fatalf("can't parse %s", URL)
	}
	select {
	case <-ctx.Done():
		log.Fatalf("can't access %s", URL)
	case <-waitForReady(net.JoinHostPort(u.Hostname(), u.Port())):
	}
}

func getEnvOrFail(s string) string {
	res := os.Getenv(s)
	if res == "" {
		log.Fatalf("no env '%s'", s)
	}
	return res
}

func waitForReady(url string) <-chan struct{} {
	res := make(chan struct{}, 1)
	go func() {
		for {
			if err := listen(url); err != nil {
				log.Printf("waiting for %s ...", url)
				time.Sleep(500 * time.Millisecond)
			} else {
				res <- struct{}{}
				return
			}
		}
	}()
	return res
}

func listen(urlStr string) error {
	log.Printf("dial %s", urlStr)
	conn, err := net.DialTimeout("tcp", urlStr, time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	return err
}

func TestLive(t *testing.T) {
	t.Parallel()
	checkCode(t, invoke(t, newRequest(t, http.MethodGet, "/live", "")), http.StatusOK)
}

func TestPost_Success(t *testing.T) {
	t.Parallel()
	resp := invoke(t, newRequest(t, http.MethodPost, "/tag", "Olia"))
	checkCode(t, resp, http.StatusOK)
	res := []service.ResultWord{}
	decode(t, resp, &res)
	require.NotEmpty(t, res)
	require.Equal(t, 2, len(res))
	assert.Equal(t, service.ResultWord{Type: "WORD", String: "Olia", Mi: "Ig", Lemma: "olia"}, res[0])
	assert.Equal(t, service.ResultWord{Type: "SENTENCE_END"}, res[1])
}

func TestPost_Number(t *testing.T) {
	t.Parallel()
	resp := invoke(t, newRequest(t, http.MethodPost, "/tag", "12"))
	checkCode(t, resp, http.StatusOK)
	res := []service.ResultWord{}
	decode(t, resp, &res)
	require.NotEmpty(t, res)
	require.Equal(t, 2, len(res))
	assert.Equal(t, service.ResultWord{Type: "NUMBER", String: "12", Mi: "M----d-", Lemma: ""}, res[0])
	assert.Equal(t, service.ResultWord{Type: "SENTENCE_END"}, res[1])
}

func TestPost_Empty(t *testing.T) {
	t.Parallel()
	resp := invoke(t, newRequest(t, http.MethodPost, "/tag", ""))
	checkCode(t, resp, http.StatusBadRequest)
}

func newRequest(t *testing.T, method string, urlSuffix string, body string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, cfg.url+urlSuffix, strings.NewReader(body))
	require.Nil(t, err, "not nil error = %v", err)
	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return req
}

func invoke(t *testing.T, r *http.Request) *http.Response {
	t.Helper()
	resp, err := cfg.httpclient.Do(r)
	require.Nil(t, err, "not nil error = %v", err)
	return resp
}

func checkCode(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		b, _ := ioutil.ReadAll(resp.Body)
		require.Equal(t, expected, resp.StatusCode, string(b))
	}
}

func decode(t *testing.T, resp *http.Response, to interface{}) {
	t.Helper()
	require.Nil(t, json.NewDecoder(resp.Body).Decode(to))
}
