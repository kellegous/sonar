package main

import (
	"flag"
	"log"
	"time"

	"sonar/config"
	"sonar/ping"
	"sonar/store"
)

func monitor(cfg *config.Config, s *store.Store) {
	for {
		for _, host := range cfg.Hosts {
			p, err := ping.NewPinger()
			if err != nil {
				log.Panic(err)
			}

			res := make([]time.Duration, cfg.SamplesPerPeriod)
			for i := 0; i < cfg.SamplesPerPeriod; i++ {
				res[i], _ = p.Ping(host.IP, i)
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

	ch := make(chan struct{})
	<-ch
}
