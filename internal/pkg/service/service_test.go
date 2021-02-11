package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/airenas/lt-pos-tagger/internal/pkg/api"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var (
	tData *Data
	tEcho *echo.Echo
	tResp *httptest.ResponseRecorder
)

func initTest(t *testing.T) {
	sr := &api.SegmenterResult{Seg: [][]int{{0, 4}, {5, 1}}, S: [][]int{{0, 6}}}
	tr := &api.TaggerResult{Msd: [][][]string{{{"xxxx", "mama"}, {"xxxx", "."}}, {{"xxx", "."}}}}
	tData = &Data{Tagger: &testTagger{res: tr}, Segmenter: &testLex{res: sr}}
	tEcho = initRoutes(tData)
	tResp = httptest.NewRecorder()
}

func TestLive(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest(http.MethodGet, "/live", nil)

	tEcho.ServeHTTP(tResp, req)
	assert.Equal(t, http.StatusOK, tResp.Code)
	assert.Equal(t, `{"service":"OK"}`, tResp.Body.String())
}

func TestNotFound(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest(http.MethodGet, "/any", strings.NewReader(``))
	req.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)

	tEcho.ServeHTTP(tResp, req)

	assert.Equal(t, http.StatusNotFound, tResp.Code)
}

func TestProvides(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest(http.MethodPost, "/tag", strings.NewReader("mama o"))

	tEcho.ServeHTTP(tResp, req)

	assert.Equal(t, http.StatusOK, tResp.Code)
	assert.Equal(t, `[{"type":"WORD","string":"mama","mi":"mama","lemma":"xxxx"},{"type":"SPACE","string":" "},{"type":"WORD","string":"o","mi":".","lemma":"xxx"},{"type":"SENTENCE_END"}]`,
		strings.TrimSpace(tResp.Body.String()))

}

func TestFails_Empty(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/tag", strings.NewReader(""))

	tEcho.ServeHTTP(tResp, req)

	assert.Equal(t, http.StatusBadRequest, tResp.Code)
}

func TestFails_Map(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/tag", strings.NewReader("mama"))

	tEcho.ServeHTTP(tResp, req)

	assert.Equal(t, http.StatusInternalServerError, tResp.Code)
}

func TestFailsMorph(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/tag", strings.NewReader("mama o"))

	tData.Tagger = &testTagger{err: errors.New("err")}
	tEcho.ServeHTTP(tResp, req)

	assert.Equal(t, http.StatusInternalServerError, tResp.Code)
}

func TestFailsLex(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/tag", strings.NewReader("mama o"))

	tData.Segmenter = &testLex{err: errors.New("err")}
	tEcho.ServeHTTP(tResp, req)

	assert.Equal(t, http.StatusInternalServerError, tResp.Code)
}

func TestMapOK(t *testing.T) {
	sr := &api.SegmenterResult{Seg: [][]int{{0, 4}}, S: [][]int{{0, 4}}}
	tr := &api.TaggerResult{Msd: [][][]string{{{"mama", "xxxx"}}}}
	r, err := MapRes("mami", tr, sr)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(r))
	assert.Equal(t, "WORD", r[0].Type)
	assert.Equal(t, "mami", r[0].String)
	assert.Equal(t, "mama", r[0].Lemma)
	assert.Equal(t, "xxxx", r[0].Mi)
}

func TestMapSeveral(t *testing.T) {
	sr := &api.SegmenterResult{Seg: [][]int{{0, 4}, {5, 2}}, S: [][]int{{0, 7}}}
	tr := &api.TaggerResult{Msd: [][][]string{{{"mama", "xxxx"}}, {{"oo", "xoo"}}}}
	r, err := MapRes("mami oi", tr, sr)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(r))
	assert.Equal(t, "WORD", r[0].Type)
	assert.Equal(t, "mami", r[0].String)
	assert.Equal(t, "mama", r[0].Lemma)
	assert.Equal(t, "xxxx", r[0].Mi)

	assert.Equal(t, "SPACE", r[1].Type)
	assert.Equal(t, " ", r[1].String)

	assert.Equal(t, "WORD", r[2].Type)
	assert.Equal(t, "oi", r[2].String)
	assert.Equal(t, "oo", r[2].Lemma)
	assert.Equal(t, "xoo", r[2].Mi)
}

func TestMapUTF(t *testing.T) {
	sr := &api.SegmenterResult{Seg: [][]int{{0, 4}, {5, 2}}, S: [][]int{{0, 7}}}
	tr := &api.TaggerResult{Msd: [][][]string{{{"mama", "xxxx"}}, {{"oo", "xoo"}}}}
	r, err := MapRes("mamą oš", tr, sr)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(r))
	assert.Equal(t, "WORD", r[0].Type)
	assert.Equal(t, "mamą", r[0].String)
	assert.Equal(t, "mama", r[0].Lemma)
	assert.Equal(t, "xxxx", r[0].Mi)

	assert.Equal(t, "SPACE", r[1].Type)

	assert.Equal(t, "WORD", r[2].Type)
	assert.Equal(t, "oš", r[2].String)
	assert.Equal(t, "oo", r[2].Lemma)
	assert.Equal(t, "xoo", r[2].Mi)
}

func TestMapSep(t *testing.T) {
	sr := &api.SegmenterResult{Seg: [][]int{{0, 1}}, S: [][]int{{0, 1}}}
	tr := &api.TaggerResult{Msd: [][][]string{{{".", "T."}}}}
	r, err := MapRes(".", tr, sr)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(r))
	assert.Equal(t, "SEPARATOR", r[0].Type)
	assert.Equal(t, ".", r[0].String)
	assert.Equal(t, "", r[0].Lemma)
	assert.Equal(t, "T.", r[0].Mi)
}

func TestMapSpace(t *testing.T) {
	sr := &api.SegmenterResult{Seg: [][]int{{0, 1}, {6, 1}}, S: [][]int{{0, 7}}}
	tr := &api.TaggerResult{Msd: [][][]string{{{".", "T."}}, {{".", "T."}}}}
	r, err := MapRes(".  \n \n.", tr, sr)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(r))
	assert.Equal(t, "SPACE", r[1].Type)
	assert.Equal(t, "  \n \n", r[1].String)
}

func TestMapSpaceDash(t *testing.T) {
	sr := &api.SegmenterResult{Seg: [][]int{{0, 1}, {4, 1}}, S: [][]int{{0, 5}}}
	tr := &api.TaggerResult{Msd: [][][]string{{{".", "T."}}, {{".", "T."}}}}
	r, err := MapRes(". - .", tr, sr)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(r))
	assert.Equal(t, "SPACE", r[1].Type)
	assert.Equal(t, " - ", r[1].String)
}

func TestMapDash(t *testing.T) {
	sr := &api.SegmenterResult{Seg: [][]int{{0, 1}, {1, 1}}, S: [][]int{{0, 2}}}
	tr := &api.TaggerResult{Msd: [][][]string{{{"a", "X"}}, {{"-", "T-"}}}}
	r, err := MapRes("a-", tr, sr)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(r))
	assert.Equal(t, "SEPARATOR", r[1].Type)
	assert.Equal(t, "-", r[1].String)
}

func TestMapNumber(t *testing.T) {
	sr := &api.SegmenterResult{Seg: [][]int{{0, 4}}, S: [][]int{{0, 4}}}
	tr := &api.TaggerResult{Msd: [][][]string{{{"1234", "M----d-"}}}}
	r, err := MapRes("1234", tr, sr)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(r))
	assert.Equal(t, "NUMBER", r[0].Type)
	assert.Equal(t, "1234", r[0].String)
	assert.Equal(t, "M----d-", r[0].Mi)
}

func TestMapSentence(t *testing.T) {
	sr := &api.SegmenterResult{Seg: [][]int{{0, 4}}, S: [][]int{{0, 4}}}
	tr := &api.TaggerResult{Msd: [][][]string{{{"1234", "M----d-"}}}}
	r, err := MapRes("1234", tr, sr)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(r))
	assert.Equal(t, "SENTENCE_END", r[1].Type)
}

func TestMapSentenceSeveral(t *testing.T) {
	sr := &api.SegmenterResult{Seg: [][]int{{0, 4}, {5, 5}}, S: [][]int{{0, 4}, {5, 5}}}
	tr := &api.TaggerResult{Msd: [][][]string{{{"1234", "M----d-"}}, {{"12345", "M----d-"}}}}
	r, err := MapRes("1234 12345", tr, sr)
	assert.Nil(t, err)
	assert.Equal(t, 5, len(r))
	assert.Equal(t, "SENTENCE_END", r[1].Type)
	assert.Equal(t, "SENTENCE_END", r[4].Type)
}

func TestMapErrTooLongSeg(t *testing.T) {
	sr := &api.SegmenterResult{Seg: [][]int{{0, 4}}, S: [][]int{{0, 4}}}
	tr := &api.TaggerResult{Msd: [][][]string{{{"1234", "M----d-"}}}}
	_, err := MapRes("123", tr, sr)
	assert.NotNil(t, err)
}

func TestMapErrWrongSeg(t *testing.T) {
	sr := &api.SegmenterResult{Seg: [][]int{{0, -1}}, S: [][]int{{0, 4}}}
	tr := &api.TaggerResult{Msd: [][][]string{{{"1234", "M----d-"}}}}
	_, err := MapRes("123", tr, sr)
	assert.NotNil(t, err)
	sr = &api.SegmenterResult{Seg: [][]int{{0, 0}}}
	_, err = MapRes("123", tr, sr)
	assert.NotNil(t, err)
	sr = &api.SegmenterResult{Seg: [][]int{{0}}}
	_, err = MapRes("123", tr, sr)
	assert.NotNil(t, err)
	sr = &api.SegmenterResult{Seg: [][]int{{0, 1}, {0, 2}}}
	_, err = MapRes("1234", tr, sr)
	assert.NotNil(t, err)
}

func TestMapErrWrongMorph(t *testing.T) {
	sr := &api.SegmenterResult{Seg: [][]int{{0, 4}}, S: [][]int{{0, 4}}}
	tr := &api.TaggerResult{Msd: [][][]string{{{"1234"}}}}
	_, err := MapRes("1234", tr, sr)
	assert.NotNil(t, err)
	tr = &api.TaggerResult{Msd: [][][]string{{{}}}}
	_, err = MapRes("1234", tr, sr)
	assert.NotNil(t, err)
	sr = &api.SegmenterResult{Seg: [][]int{{0, 4}, {5, 1}}}
	tr = &api.TaggerResult{Msd: [][][]string{{{"1234", "xx"}}}}
	_, err = MapRes("1234 .", tr, sr)
	assert.NotNil(t, err)
}

func TestMapSentence_Error(t *testing.T) {
	sr := &api.SegmenterResult{Seg: [][]int{{0, 4}}, S: [][]int{{}}}
	tr := &api.TaggerResult{Msd: [][][]string{{{"1234", "M----d-"}}}}
	_, err := MapRes("1234", tr, sr)
	assert.NotNil(t, err)
}

func TestMapSentence_ErrorWrongSentence(t *testing.T) {
	sr := &api.SegmenterResult{Seg: [][]int{{0, 4}}, S: [][]int{{1, 0}}}
	tr := &api.TaggerResult{Msd: [][][]string{{{"1234", "M----d-"}}}}
	_, err := MapRes("1234", tr, sr)
	assert.NotNil(t, err)
}

func TestMapSentence_ErrorNoSentence(t *testing.T) {
	sr := &api.SegmenterResult{Seg: [][]int{{0, 4}, {5, 1}}, S: [][]int{{0, 4}}}
	tr := &api.TaggerResult{Msd: [][][]string{{{"1234", "M----d-"}}, {{"1", "M----d-"}}}}
	_, err := MapRes("1234 1", tr, sr)
	assert.NotNil(t, err)
}

type testTagger struct {
	res *api.TaggerResult
	err error
}

func (s *testTagger) Process(string, *api.SegmenterResult) (*api.TaggerResult, error) {
	return s.res, s.err
}

type testLex struct {
	res *api.SegmenterResult
	err error
}

func (s *testLex) Process(string) (*api.SegmenterResult, error) {
	return s.res, s.err
}

type testPreprocessor struct {
	err error
}

func (s *testPreprocessor) Process(text string) (string, error) {
	return text, s.err
}
