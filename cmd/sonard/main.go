package main

import (
	"context"
	"flag"
	"hash/fnv"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/kellegous/glue/build"
	"github.com/kellegous/glue/devmode"
	"github.com/kellegous/glue/logging"
	"go.uber.org/zap"

	"github.com/kellegous/sonar/internal/config"
	"github.com/kellegous/sonar/internal/ping"
	"github.com/kellegous/sonar/internal/store"
	"github.com/kellegous/sonar/internal/ui"
	"github.com/kellegous/sonar/internal/web"
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

func getAssets(ctx context.Context, devMode *devmode.Flag) (http.Handler, error) {
	if !devMode.IsEnabled() {
		return ui.Assets()
	}

	return devmode.AssetsFromVite(
		ctx,
		devMode,
		devmode.WithBuildSummary(build.ReadSummary()),
		devmode.UseBun())
}

func main() {
	var flags struct {
		ConfigFile string
		DevMode    devmode.Flag
	}

	flag.StringVar(
		&flags.ConfigFile,
		"conf",
		"sonar.toml",
		"the config file for the service")
	flag.Var(
		&flags.DevMode,
		"dev-mode",
		"Enable dev mode")
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

	assets, err := getAssets(ctx, &flags.DevMode)
	if err != nil {
		lg.Fatal("unable to load assets",
			zap.Error(err))
	}

	go monitor(&cfg, s)

	go func() {
		ctx, done := context.WithTimeout(ctx, 30*time.Second)
		defer done()
		if err := flags.DevMode.ShowBannerWhenReady(
			ctx,
			os.Stdout,
			cfg.Addr,
		); err != nil {
			lg.Fatal("unable to show banner",
				zap.Error(err))
		}
	}()

	if err := web.ListenAndServe(ctx, &cfg, s, assets); err != nil {
		lg.Fatal("unable to serve web traffic",
			zap.Error(err))
	}
}
