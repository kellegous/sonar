package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"

	"github.com/kellegous/glue/build"
)

const defaultVitePort = 3001

func startVite(
	ctx context.Context,
	root string,
	port int,
) error {
	bs := build.ReadSummary()

	c := exec.CommandContext(
		ctx,
		"node_modules/.bin/vite",
		"--clearScreen=false",
		fmt.Sprintf("--port=%d", port))
	c.Dir = root
	c.Env = append(
		os.Environ(),
		fmt.Sprintf("SHA=%s", bs.SHA),
		fmt.Sprintf("BUILD_NAME=%s", bs.Name),
	)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Start()
}

func startSonar(
	ctx context.Context,
	root string,
	proxyURL string,
) error {
	c := exec.CommandContext(
		ctx,
		"sudo",
		"bin/sonard",
		fmt.Sprintf("--web.asset-proxy-url=%s", proxyURL))
	c.Dir = root
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Start()
}

func main() {
	var flags struct {
		Root string
		Vite struct {
			Port int
		}
	}

	flag.StringVar(
		&flags.Root,
		"root",
		".",
		"Root directory for the server")

	flag.IntVar(
		&flags.Vite.Port,
		"web.vite-port",
		defaultVitePort,
		"the port for the vite server")

	flag.Parse()

	ctx, done := signal.NotifyContext(
		context.Background(),
		os.Interrupt)
	defer done()

	if err := startVite(
		ctx,
		flags.Root,
		flags.Vite.Port,
	); err != nil {
		log.Panic(err)
	}

	if err := startSonar(
		ctx,
		flags.Root,
		fmt.Sprintf("http://localhost:%d", flags.Vite.Port),
	); err != nil {
		log.Panic(err)
	}

	<-ctx.Done()
}
