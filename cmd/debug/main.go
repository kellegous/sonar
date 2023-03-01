package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/kellegous/sonar/pkg/config"
	"github.com/kellegous/sonar/pkg/store"
)

func toInts(durs []time.Duration) string {
	v := make([]string, 0, len(durs))
	for _, dur := range durs {
		v = append(v, fmt.Sprintf("%d (%x)", int64(dur), int64(dur)))
	}
	return strings.Join(v, ",")
}

func main() {
	flagConf := flag.String("config", "sonar.toml", "")
	flag.Parse()

	var cfg config.Config
	if err := cfg.ReadFile(*flagConf); err != nil {
		log.Panic(err)
	}

	s, err := store.Open(cfg.DataPath)
	if err != nil {
		log.Panic(err)
	}

	for _, host := range cfg.Hosts {
		if err := s.ForEach(
			store.NewMarker(host.IP, store.First),
			store.NewMarker(host.IP, store.Last),
			func(ip net.IP, t time.Time, r []time.Duration) error {
				for _, x := range r {
					if x < 0 {
						fmt.Printf("%s, %s, %s\n", ip, t, toInts(r))
						break
					}
				}
				return nil
			}); err != nil {
			log.Panic(err)
		}
	}
}
