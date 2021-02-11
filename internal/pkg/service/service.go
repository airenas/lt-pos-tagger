package service

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/lt-pos-tagger/internal/pkg/api"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

type taggerHandler struct {
	data *Data
}

func (t *taggerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

//StartWebServer starts the HTTP service and listens for the requests
func StartWebServer(data *Data) error {
	goapp.Log.Infof("Starting HTTP service at %d", data.Port)
	portStr := strconv.Itoa(data.Port)

	e := initRoutes(data)

	e.Server.Addr = ":" + portStr

	w := goapp.Log.Writer()
	defer w.Close()
	l := log.New(w, "", 0)
	gracehttp.SetLogger(l)

	return gracehttp.Serve(e.Server)
}

func initRoutes(data *Data) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	p := prometheus.NewPrometheus("acronyms", nil)
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

func (cb *textBinder) Bind(s *string, c echo.Context) error {
	bodyBytes, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Can't get data").SetInternal(err)
	}
	*s = string(bodyBytes)
	*s = strings.TrimSpace(string(bodyBytes))
	if *s == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "No input", err)
	}
	return nil
}

func handleText(data *Data) func(echo.Context) error {
	return func(c echo.Context) error {
		defer goapp.Estimate("Service method: tag")()
		tb := &textBinder{}
		var text string
		if err := tb.Bind(&text, c); err != nil {
			goapp.Log.Error(err)
			return err
		}

		sgm, err := data.Segmenter.Process(text)
		if err != nil {
			goapp.Log.Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Can't segment")
		}
		goapp.Log.Debugf("Lex: %v", sgm)

		tgr, err := data.Tagger.Process(text, sgm)
		if err != nil {
			goapp.Log.Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Can't tag")
		}
		logrus.Debugf("Tagger: %v", tgr)

		res, err := MapRes(text, tgr, sgm)
		if err != nil {
			goapp.Log.Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Can't map")
		}
		logrus.Debugf("Res: %v", res)

		return c.JSON(http.StatusOK, res)
	}
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
			return nil, errors.Errorf("Wrong seg %v", s)
		}
		if s[0] < 0 || s[1] < 1 {
			return nil, errors.Errorf("Wrong seg %v", s)
		}
		if s[0]+s[1] > len(rns) {
			return nil, errors.Errorf("Wrong seg %v", s)
		}
		if sent == nil {
			return nil, errors.Errorf("No sentence for %v", s)
		}
		t := string(rns[s[0] : s[0]+s[1]])
		if len(tgr.Msd) <= i {
			return nil, errors.Errorf("No msd at %d", i)
		}
		if len(tgr.Msd[i]) < 1 {
			return nil, errors.Errorf("No msd at %d", i)
		}
		if len(tgr.Msd[i][0]) < 2 {
			return nil, errors.Errorf("No msd at %d", i)
		}
		mi := tgr.Msd[i][0][1]
		if ep < s[0] {
			res = append(res, space(string(rns[ep:s[0]]))...)
		}
		if isSep(mi) {
			res = append(res, sep(t, mi))
		} else if isNum(mi) {
			res = append(res, num(t, mi))
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

func space(s string) []ResultWord {
	strs := strings.Split(s, "-")
	res := make([]ResultWord, 0)
	skip := false
	for _, sp := range strs {
		if sp == "" {
			res = append(res, ResultWord{Type: "SEPARATOR", String: "-"})
			skip = true
		} else {
			if len(res) > 0 && !skip {
				res = append(res, ResultWord{Type: "SEPARATOR", String: "-"})
			}
			res = append(res, ResultWord{Type: "SPACE", String: sp})
			skip = false
		}
	}
	return res
}

func sep(s string, mi string) ResultWord {
	return ResultWord{Type: "SEPARATOR", String: s, Mi: mi}
}

func sentenceEnd() ResultWord {
	return ResultWord{Type: "SENTENCE_END"}
}

func num(s string, mi string) ResultWord {
	return ResultWord{Type: "NUMBER", String: s, Mi: mi}
}

func word(s, mf, mi string) ResultWord {
	return ResultWord{Type: "WORD", String: s, Lemma: mf, Mi: mi}
}

func isSep(mi string) bool {
	return strings.HasPrefix(mi, "T")
}

func isNum(mi string) bool {
	return mi == "M----rn" || mi == "M----d-"
}
