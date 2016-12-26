package web

import (
	"encoding/json"
	"log"
	"math"
	"net"
	"net/http"
	"sonar/config"
	"sonar/store"
	"strconv"
	"strings"
	"time"

	"github.com/kellegous/pork"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Panic(err)
	}
}

func writeJSONOk(w http.ResponseWriter, data interface{}) {
	r := struct {
		Ok   bool        `json:"ok"`
		Data interface{} `json:"data"`
	}{
		true,
		data,
	}
	writeJSON(w, http.StatusOK, &r)
}

func writeJSONErr(w http.ResponseWriter, status int, err error) {
	r := struct {
		Ok  bool   `json:"ok"`
		Err string `json:"error"`
	}{
		false,
		err.Error(),
	}
	writeJSON(w, status, &r)
}

type current struct {
	IP   string    `json:"ip"`
	Name string    `json:"name"`
	Time time.Time `json:"time"`
	Summary
}

type hour struct {
	Time time.Time `json:"time"`
	Summary
}

type hourly struct {
	IP    string  `json:"ip"`
	Name  string  `json:"name"`
	Hours []*hour `json:"hours"`
}

// Summary ...
type Summary struct {
	Avg    float64 `json:"avg"`
	Stddev float64 `json:"stddev"`
	Max    int     `json:"max"`
	Min    int     `json:"min"`
	Count  int     `json:"count"`
	Data   []int   `json:"data,omitempty"`
}

func summarize(s *Summary, vals []time.Duration, withData bool) {
	*s = Summary{}

	n := len(vals)
	if n == 0 {
		return
	}

	var max int64
	var min int64 = 0x7fffffffffffffff

	mu := 0.0
	for _, d := range vals {
		mu += float64(d.Nanoseconds())
		if d.Nanoseconds() > max {
			max = d.Nanoseconds()
		}
		if d.Nanoseconds() < min {
			min = d.Nanoseconds()
		}
	}
	mu /= float64(n)

	std := 0.0
	for _, d := range vals {
		x := float64(d.Nanoseconds()) - mu
		std += x * x
	}
	std = math.Sqrt(std / float64(n))

	s.Avg = mu
	s.Stddev = std
	s.Max = int(max)
	s.Min = int(min)
	s.Count = n

	if withData {
		for _, d := range vals {
			s.Data = append(s.Data, int(d.Nanoseconds()))
		}
	}
}

func apiCurrent(cfg *config.Config, s *store.Store, w pork.ResponseWriter, r *http.Request) {
	withRaw := boolParam(r.FormValue("with-raw"))
	res := make([]*current, 0, len(cfg.Hosts))
	for _, host := range cfg.Hosts {
		t, vals, err := s.Current(host.IP)
		if err != nil {
			log.Panic(err)
		}

		c := &current{
			IP:   host.IP.String(),
			Name: host.Name,
			Time: t,
		}

		summarize(&c.Summary, vals, withRaw)
		res = append(res, c)
	}

	writeJSONOk(w, res)
}

func boolParam(v string) bool {
	v = strings.ToLower(v)
	return v == "true" || v == "yes" || v == "yep"
}

func intParam(v string, def int) int {
	if v == "" {
		return def
	}

	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}

	return int(i)
}

func apiByHour(cfg *config.Config, s *store.Store, w pork.ResponseWriter, r *http.Request) {
	hrs := intParam(r.FormValue("n"), 24)
	withRaw := boolParam(r.FormValue("with-raw"))

	st := time.Now().Add(-time.Duration(hrs) * time.Hour).Truncate(time.Hour)

	var res []*hourly

	for _, host := range cfg.Hosts {
		data := make([][]time.Duration, hrs)
		if err := s.ForEach(
			store.NewMarker(host.IP, st),
			store.NewMarker(host.IP, store.Last),
			func(ip net.IP, t time.Time, vals []time.Duration) error {
				ix := int(t.Sub(st).Nanoseconds() / int64(time.Hour))
				if ix < 0 || ix >= hrs {
					return nil
				}
				for _, val := range vals {
					data[ix] = append(data[ix], val)
				}
				return nil
			}); err != nil {
			log.Panic(err)
		}

		curr := &hourly{
			IP:    host.IP.String(),
			Name:  host.Name,
			Hours: make([]*hour, 0, hrs),
		}

		for ix, vals := range data {
			h := &hour{Time: st.Add(time.Duration(ix) * time.Hour)}
			summarize(&h.Summary, vals, withRaw)
			curr.Hours = append(curr.Hours, h)
		}

		res = append(res, curr)
	}

	writeJSONOk(w, res)
}

func setupAPI(r pork.Router, cfg *config.Config, s *store.Store) {
	r.RespondWithFunc(
		"/api/v1/current",
		func(w pork.ResponseWriter, r *http.Request) {
			apiCurrent(cfg, s, w, r)
		})

	r.RespondWithFunc(
		"/api/v1/hourly",
		func(w pork.ResponseWriter, r *http.Request) {
			apiByHour(cfg, s, w, r)
		})
}
