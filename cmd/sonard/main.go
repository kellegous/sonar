package main

import (
	"context"
	"flag"
	"hash/fnv"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

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

func proxyTo(u *url.URL) http.Handler {
	p := httputil.NewSingleHostReverseProxy(u)
	dir := p.Director
	p.Director = func(r *http.Request) {
		dir(r)
		r.Host = u.Host
	}
	return p
}

func getAssets(proxyURL string) (http.Handler, error) {
	if proxyURL == "" {
		return ui.Assets()
	}

	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	return proxyTo(u), nil
}

func main() {
	var flags struct {
		ConfigFile string
		Web        struct {
			AssetProxyURL string
		}
	}

	flag.StringVar(
		&flags.ConfigFile,
		"conf",
		"sonar.toml",
		"the config file for the service")

	flag.StringVar(
		&flags.Web.AssetProxyURL,
		"web.asset-proxy-url",
		"",
		"the URL to use for the asset proxy")
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

	assets, err := getAssets(flags.Web.AssetProxyURL)
	if err != nil {
		lg.Fatal("unable to load assets",
			zap.Error(err))
	}

	go monitor(&cfg, s)

	svr := &web.Server{
		Config: &cfg,
		Store:  s,
	}

	if err := svr.ListenAndServe(ctx, assets); err != nil {
		lg.Fatal("unable to serve web traffic",
			zap.Error(err))
	}
}
