package service

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/lt-pos-tagger/internal/pkg/api"
	"github.com/airenas/lt-pos-tagger/internal/pkg/utils"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"

	"github.com/pkg/errors"
)

type (
	// Tagger returns word forms
	Tagger interface {
		Process(string, *api.SegmenterResult) (*api.TaggerResult, error)
	}

	//Segmenter segments text
	Segmenter interface {
		Process(text string) (*api.SegmenterResult, error)
	}

	//Data is service operation data
	Data struct {
		Tagger    Tagger
		Segmenter Segmenter
		Port      int
	}
)

//StartWebServer starts the HTTP service and listens for the requests
func StartWebServer(data *Data) error {
	goapp.Log.Infof("Starting HTTP service at %d", data.Port)
	portStr := strconv.Itoa(data.Port)

	e := initRoutes(data)

	e.Server.Addr = ":" + portStr
	e.Server.IdleTimeout = 3 * time.Minute
	e.Server.ReadHeaderTimeout = 10 * time.Second
	e.Server.ReadTimeout = 20 * time.Second
	e.Server.WriteTimeout = 30 * time.Second

	w := goapp.Log.Writer()
	defer w.Close()
	l := log.New(w, "", 0)
	gracehttp.SetLogger(l)

	return gracehttp.Serve(e.Server)
}

func initRoutes(data *Data) *echo.Echo {
	e := echo.New()
	p := prometheus.NewPrometheus("tag", nil)
	p.Use(e)

	e.POST("/tag", handleText(data))
	e.GET("/live", live(data))

	goapp.Log.Info("Routes:")
	for _, r := range e.Routes() {
		goapp.Log.Infof("  %s %s", r.Method, r.Path)
	}
	return e
}

type textBinder struct{}

func (cb *textBinder) Bind(c echo.Context, s *string) error {
	bodyBytes, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Can't get data").SetInternal(err)
	}
	*s = strings.TrimSpace(string(bodyBytes))
	if *s == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "No input")
	}
	return nil
}

func handleText(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: tag")()
		tb := &textBinder{}
		var text string
		if err := tb.Bind(c, &text); err != nil {
			goapp.Log.Error(err)
			return err
		}

		sgm, err := data.Segmenter.Process(text)
		if err != nil {
			goapp.Log.Error(err)
			return echo.NewHTTPError(mapHTTPError(err), "Can't segment")
		}

		tgr, err := data.Tagger.Process(text, sgm)
		if err != nil {
			goapp.Log.Error(err)
			return echo.NewHTTPError(mapHTTPError(err), "Can't tag")
		}
		goapp.Log.Debugf("Tagger: %v", tgr)

		res, err := MapRes(text, tgr, sgm)
		if err != nil {
			goapp.Log.Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Can't map")
		}
		goapp.Log.Debugf("Res: %v", res)

		return c.JSON(http.StatusOK, res)
	}
}

func mapHTTPError(err error) int {
	if err == utils.ErrTooBusy {
		return http.StatusTooManyRequests
	}
	return http.StatusInternalServerError
}

func live(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		return c.JSONBlob(http.StatusOK, []byte(`{"service":"OK"}`))
	}
}

//MapRes map function
func MapRes(text string, tgr *api.TaggerResult, sgm *api.SegmenterResult) ([]ResultWord, error) {
	res := make([]ResultWord, 0)
	si := 0
	ep := 0
	rns := []rune(text)
	sent := getSentence(sgm.S, si)
	for i, s := range sgm.Seg {
		if len(s) < 2 {
			return nil, errors.Errorf("Wrong seg (< 2) %v", s)
		}
		if s[0] < 0 || s[1] < 1 {
			return nil, errors.Errorf("Wrong seg %v", s)
		}
		if s[0]+s[1] > len(rns) {
			return nil, errors.Errorf("Wrong seg (len > len(s)) %v, %d. %s", s, len(rns), tryTakeText(rns, s[0]))
		}
		if sent == nil {
			return nil, errors.Errorf("No sentence for %v", s)
		}
		t := string(rns[s[0] : s[0]+s[1]])
		if len(tgr.Msd) <= i {
			return nil, errors.Errorf("No msd at %d. %s", i, tryTakeText(rns, s[0]))
		}
		if len(tgr.Msd[i]) < 1 {
			return nil, errors.Errorf("Wrong msd at (len < 1) %d. %s", i, tryTakeText(rns, s[0]))
		}
		if len(tgr.Msd[i][0]) < 2 {
			return nil, errors.Errorf("Wrong msd at (len[0] < 2) %d. %s", i, tryTakeText(rns, s[0]))
		}
		mi := tgr.Msd[i][0][1]
		if ep < s[0] {
			res = append(res, space(string(rns[ep:s[0]])))
		}

		if isNum(t, mi) {
			res = append(res, num(t, mi))
		} else if isSep(mi) {
			res = append(res, sep(t, mi))
		} else {
			res = append(res, word(t, tgr.Msd[i][0][0], mi))
		}
		ep = s[0] + s[1]
		if ep >= (sent[0] + sent[1]) {
			res = append(res, sentenceEnd())
			si++
			sent = getSentence(sgm.S, si)
		}
	}
	return res, nil
}

func getSentence(s [][]int, i int) []int {
	if len(s) <= i {
		return nil
	}
	res := s[i]
	if len(res) < 2 {
		return nil
	}
	if res[0] < 0 || res[1] < 1 {
		return nil
	}
	return res
}

func tryTakeText(rns []rune, from int) string {
	return string(rns[max(from-10, 0):min(from+100, len(rns))])
}

func min(i1, i2 int) int {
	if i1 > i2 {
		return i2
	}
	return i1
}

func max(i1, i2 int) int {
	if i1 < i2 {
		return i2
	}
	return i1
}

func space(s string) ResultWord {
	return ResultWord{Type: "SPACE", String: s}
}

func sep(s string, mi string) ResultWord {
	return ResultWord{Type: "SEPARATOR", String: s, Mi: mi}
}

func sentenceEnd() ResultWord {
	return ResultWord{Type: "SENTENCE_END"}
}

func num(s string, mi string) ResultWord {
	tmi := mi
	if mi == "Th" || mi == "X-" {
		tmi = "M----d-" // negative number workaround
	}
	return ResultWord{Type: "NUMBER", String: s, Mi: tmi}
}

func word(s, mf, mi string) ResultWord {
	return ResultWord{Type: "WORD", String: s, Lemma: mf, Mi: mi}
}

func isSep(mi string) bool {
	return strings.HasPrefix(mi, "T")
}

func isNum(s string, mi string) bool {
	return mi == "M----rn" || mi == "M----d-" ||
		((mi == "Th" || mi == "X-") && len(s) > 1 && utils.IsNumber(s)) // negative number workaround
}
