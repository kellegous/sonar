package main

import (
	"context"
	"flag"
	"hash/fnv"
	"log"
	"net"
	"time"

	"go.uber.org/zap"

	"github.com/kellegous/sonar/pkg/config"
	"github.com/kellegous/sonar/pkg/logging"
	"github.com/kellegous/sonar/pkg/ping"
	"github.com/kellegous/sonar/pkg/store"
	"github.com/kellegous/sonar/pkg/web"
)

func idFor(ip net.IP) int {
	h := fnv.New32()
	h.Write(ip)
	return int(h.Sum32() & 0xffff)
}

func monitor(cfg *config.Config, s *store.Store) {
	for {
		now := time.Now()
		logging.L(context.Background()).Info("pinging hosts",
			zap.Time("now", now))

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

type Flags struct {
	ConfigFile string
	DevRoot    string
}

func (f *Flags) Register(fs *flag.FlagSet) {
	fs.StringVar(
		&f.ConfigFile,
		"conf",
		"sonar.toml",
		"the config file for the service")
	fs.StringVar(
		&f.DevRoot,
		"dev-root",
		"",
		"enable rebuilding web assets for development")
}
func main() {
	var flags Flags
	flags.Register(flag.CommandLine)
	flag.Parse()

	lg := logging.MustSetup()

	ctx := context.Background()

	var cfg config.Config
	if err := cfg.ReadFile(flags.ConfigFile); err != nil {
		lg.Fatal("unable to read config",
			zap.Error(err),
			zap.String("config", flags.ConfigFile))
	}

	s, err := store.Open(cfg.DataPath)
	if err != nil {
		lg.Fatal("unable to open store",
			zap.Error(err))
	}

	go monitor(&cfg, s)

	svr := &web.Server{
		Config: &cfg,
		Store:  s,
	}

	if err := svr.ListenAndServe(ctx, &web.Options{
		UseDevRoot: flags.DevRoot,
	}); err != nil {
		lg.Fatal("unable to serve web traffic",
			zap.Error(err))
	}
}
