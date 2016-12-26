package main

import (
	"flag"
	"log"
	"time"

	"sonar/config"
	"sonar/ping"
	"sonar/store"
	"sonar/web"
)

var assetsDir string

func monitor(cfg *config.Config, s *store.Store) {
	for {
		now := time.Now()

		for _, host := range cfg.Hosts {

			p, err := ping.NewPinger()
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
	flagConf := flag.String("config", "sonar.toml",
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
