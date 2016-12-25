package web

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"sonar/config"
	"sonar/store"
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

type stats struct {
	IP     string  `json:"ip"`
	Name   string  `json:"name"`
	Avg    float64 `json:"avg"`
	StdDev float64 `json:"avg"`
	Max    int     `json:"max"`
	Min    int     `json:"min"`
	Data   []int   `json:"data,omitempty"`
}

func toStats(h *config.Host, vals []time.Duration, withRaw bool) *stats {
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
	mu /= float64(len(vals))

	std := 0.0
	for _, d := range vals {
		x := float64(d.Nanoseconds()) - mu
		std += x * x
	}
	std = math.Sqrt(std / float64(len(vals)))

	s := &stats{
		IP:     h.IP.String(),
		Name:   h.Name,
		Avg:    mu,
		StdDev: std,
		Max:    int(max),
		Min:    int(min),
	}

	if withRaw {
		for _, d := range vals {
			s.Data = append(s.Data, int(d.Nanoseconds()))
		}
	}

	return s
}

func apiCurrent(cfg *config.Config, s *store.Store, w pork.ResponseWriter, r *http.Request) {
	res := make([]*stats, 0, len(cfg.Hosts))
	for _, host := range cfg.Hosts {
		_, vals, err := s.Current(host.IP)
		if err != nil {
			log.Panic(err)
		}

		res = append(res, toStats(host, vals, true))
	}

	writeJSONOk(w, res)
}

func apiByHour(cfg *config.Config, s *store.Store, w pork.ResponseWriter, r *http.Request) {
}

func setupAPI(r pork.Router, cfg *config.Config, s *store.Store) {
	r.RespondWithFunc(
		"/api/v1/current",
		func(w pork.ResponseWriter, r *http.Request) {
			apiCurrent(cfg, s, w, r)
		})

	r.RespondWithFunc(
		"/api/v1/by-hour",
		func(w pork.ResponseWriter, r *http.Request) {
			apiByHour(cfg, s, w, r)
		})
}
