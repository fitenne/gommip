package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func ParseConfig() *Config {
	configFilePath := flag.String("config", "config.yml", "path to config file")
	flag.Parse()

	return configFrom(*configFilePath)
}

func main() {
	config := ParseConfig()
	initLogger(config)
	if err := prepareDb(config); err != nil {
		zap.S().Fatalln(err)
	}

	serve(config)
}

func serve(config *Config) {
	s := NewServer(config)

	go func() {
		zap.S().Warnln("bring up server")
		err := s.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.S().Fatalln(err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer stop()

	<-ctx.Done()
	stop()
	zap.S().Infoln("shutting down gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		zap.S().Fatalln("Server Shutdown:", err)
	}

	zap.S().Infoln("Server exiting")
}
