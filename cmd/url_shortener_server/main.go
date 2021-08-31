package main

import (
	"context"
	"github.com/spf13/pflag"
	"os"
	"url_shortener/pkg/config"
	"url_shortener/pkg/daemon"

	log "github.com/sirupsen/logrus"
)

func main() {
	var daemonCfg string

	pflag.StringVarP(&daemonCfg, "config", "c", "config.yml", "daemon config filepath")
	pflag.Parse()

	// logger
	logLevelVar, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		logLevelVar = "info"
	}
	if logLevel, err := log.ParseLevel(logLevelVar); err == nil {
		log.SetLevel(logLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, TimestampFormat: "15:04:05 2006-01-02"})
	log.SetOutput(os.Stderr)

	// config
	cfg, err := config.Load(daemonCfg)
	if err != nil {
		log.Fatalf("cannot load config file: %v", err)
	}

	// daemon
	d, err := daemon.New(context.Background(), *cfg)
	if err != nil {
		log.Fatalf("cannot init daemon: %v", err)
	}
	if err = d.Run(); err != nil {
		log.Fatalf("daemon error: %v", err)
	}
}
