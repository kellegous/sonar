package main

import (
	"flag"
	"hash/fnv"
	"log"
	"net"
	"time"

	"sonar/config"
	"sonar/ping"
	"sonar/store"
	"sonar/web"
)

var assetsDir string

func idFor(ip net.IP) int {
	h := fnv.New32()
	h.Write(ip)
	return int(h.Sum32() & 0xffff)
}

func monitor(cfg *config.Config, s *store.Store) {
	for {
		now := time.Now()

		for _, host := range cfg.Hosts {

			p, err := ping.NewPinger(idFor(host.IP))
			if err != nil {
				log.Panic(err)
			}

			res := make([]time.Duration, cfg.SamplesPerPeriod)
			for i := 0; i < cfg.SamplesPerPeriod; i++ {
				res[i], _ = p.Ping(host.IP, i)
			}

			if err := p.Close(); err != nil {
				log.Panic(err)
			}

			if err := s.Write(host.IP, now, res); err != nil {
				log.Panic(err)
			}
		}

		time.Sleep(cfg.SamplePeriod)
	}
}

func main() {
	flagConf := flag.String("conf", "sonar.toml",
		"config file")
	flag.Parse()

	var cfg config.Config
	if err := cfg.ReadFile(*flagConf); err != nil {
		log.Panic(err)
	}

	s, err := store.Open(cfg.DataPath)
	if err != nil {
		log.Panic(err)
	}

	go monitor(&cfg, s)

	log.Panic(web.ListenAndServe(&cfg, s, assetsDir))
}
