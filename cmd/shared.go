package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
)

func setupLogger(logger *logrus.Logger, verbose bool) {
	if verbose {
		logger.SetLevel(logrus.DebugLevel)
		logger.Debug("Starting in debug level")
	} else {
		logger.SetLevel(logrus.InfoLevel)
		logger.Formatter = new(logrus.JSONFormatter)
	}
}

func newExitHandlerContext(logger *logrus.Logger) context.Context {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-c
		defer cancel()
		logger.Info("shutting down")
	}()

	return ctx
}
