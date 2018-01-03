package main

import (
	"flag"

	"github.com/Sirupsen/logrus"
)

var (
	configPath string
)

func init() {
	flag.StringVar(
		&configPath,
		"config",
		"config.toml",
		"path to configuration file; supported formats are JSON, YAML, and TOML",
	)
	flag.Parse()
}

func main() {
	v, err := parseConfig(configPath)
	if err != nil {
		logrus.Fatalf("could not parse config file (%s): %s", configPath, err)
	}
	s, err := setupGRPCServer(v)
	if err != nil {
		logrus.Fatalf("failed to initialize GRPC server: %s", err)
	}
	l, err := setupNetListener(v)
	if err != nil {
		logrus.Fatalf("failed to create net.Listener: %s", err)
	}
	logrus.Infof("attempting to start server on: %s", l.Addr().String())
	if err := s.Serve(l); err != nil {
		logrus.Fatalf("server shut down due to error: %s", err)
	}
	logrus.Info("server shutting down cleanly")
}
