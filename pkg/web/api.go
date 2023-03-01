package web

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kellegous/sonar/pkg/store"
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
	LossRatio float64 `json:"lossRatio"`
	Avg       float64 `json:"avg"`
	Max       int     `json:"max"`
	Min       int     `json:"min"`
	P90       int     `json:"p90"`
	P50       int     `json:"p50"`
	P10       int     `json:"p10"`
	Count     int     `json:"count"`
	Data      []int   `json:"data,omitempty"`
}

func perc(p float64, vals []int) int {
	// interpolation cannot work with a single element.
	if len(vals) == 1 {
		return vals[0]
	}

	ix := int(float64(len(vals)) * p)

	return vals[ix]
	// get the precise location for this percentile
	// ix := float64(len(vals))*p - 0.5

	// split this into integral and fractional
	// pi, pf := math.Modf(ix)

	// interpolate the value from the two referenced locations
	// v := float64(vals[int(pi)])*(1-pf) + float64(vals[int(pi)+1])*pf

	// return int(v)
}

func summarize(s *Summary, vals []time.Duration, withData bool) {
	*s = Summary{}

	n := len(vals)
	if n == 0 {
		return
	}

	// filter out any lost packets
	valid := make([]int, 0, n)
	for _, d := range vals {
		v := int(d.Nanoseconds())
		if v == 0 {
			continue
		}
		valid = append(valid, v)
	}

	// TODO(knorton): should this be len(valid)?
	s.Count = n
	s.LossRatio = 1.0 - float64(len(valid))/float64(n)

	if withData {
		for _, d := range vals {
			s.Data = append(s.Data, int(d.Nanoseconds()))
		}
	}

	// if all packets were lost, we can't do anything else.
	if len(valid) == 0 {
		return
	}

	sort.Ints(valid)

	max := 0
	min := 0x7fffffffffffffff

	mu := 0.0
	for _, d := range valid {
		mu += float64(d)
		if d > max {
			max = d
		}
		if d < min {
			min = d
		}
	}
	mu /= float64(len(valid))

	s.Avg = mu
	s.Max = int(max)
	s.Min = int(min)
	s.P10 = perc(0.1, valid)
	s.P90 = perc(0.9, valid)
	s.P50 = perc(0.5, valid)
}

func apiCurrent(
	w http.ResponseWriter,
	r *http.Request,
	s *Server,
) {
	withRaw := boolParam(r.FormValue("with-raw"))
	cfg := s.Config
	res := make([]*current, 0, len(cfg.Hosts))
	for _, host := range cfg.Hosts {
		t, vals, err := s.Store.Current(host.IP)
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

func apiByHour(
	w http.ResponseWriter,
	r *http.Request,
	s *Server,
) {
	cfg := s.Config
	hrs := intParam(r.FormValue("n"), 24)
	withRaw := boolParam(r.FormValue("with-raw"))

	st := time.Now().Add(-time.Duration(hrs-1) * time.Hour).Truncate(time.Hour)

	var res []*hourly

	for _, host := range cfg.Hosts {
		data := make([][]time.Duration, hrs)
		if err := s.Store.ForEach(
			store.NewMarker(host.IP, st),
			store.NewMarker(host.IP, store.Last),
			func(ip net.IP, t time.Time, vals []time.Duration) error {
				ix := int(t.Sub(st).Nanoseconds() / int64(time.Hour))
				if ix < 0 || ix > hrs {
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
