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
	zapsink "github.com/kellegous/glue/logging/yarder/zap"
	"github.com/kellegous/poop"
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

type Flags struct {
	ConfigFile string
	DevMode    devmode.Flag
	Logging    LoggingFlags
}

func (f *Flags) Register(fs *flag.FlagSet) {
	fs.StringVar(
		&f.ConfigFile,
		"conf",
		"sonar.toml",
		"the config file for the service")

	fs.Var(
		&f.DevMode,
		"dev-mode",
		"Enable dev mode")

	fs.Var(
		&f.Logging.Level,
		"logging.level",
		"logging: the level to log at")
	fs.Var(
		&f.Logging.Outputs,
		"logging.output",
		"add the following output to the logging pipeline")
}

type LoggingFlags struct {
	Level   logging.LevelFlag
	Outputs logging.OutputPathsFlag
}

func main() {
	if err := zapsink.Register(zapsink.WithApp("sonar")); err != nil {
		poop.HitFan(err)
	}
	flags := Flags{
		Logging: LoggingFlags{
			Outputs: logging.NewOutputPathsFlag("stderr"),
		},
	}
	flags.Register(flag.CommandLine)
	flag.Parse()

	lg := logging.MustSetup(
		logging.WithLevel(flags.Logging.Level.Level()),
		logging.WithOutputPaths(flags.Logging.Outputs.Paths()...),
	)

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
