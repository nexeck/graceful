package http

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog/log"
)

type GracefulHTTP interface{}

type gracefulHTTP struct {
	httpServer      *http.Server
	gracefulTimeout time.Duration
}

func New(httpServer *http.Server, gracefulTimeout time.Duration) *gracefulHTTP {
	return &gracefulHTTP{
		httpServer:      httpServer,
		gracefulTimeout: gracefulTimeout,
	}
}

func (g *gracefulHTTP) Run() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	go func() {
		log.Info().Msgf("HTTP Listen Address: %s", g.httpServer.Addr)
		log.Error().Err(g.httpServer.ListenAndServe()).Msg("HTTP Server error")
	}()

	<-stopChan // wait for SIGINT
	log.Info().Msgf("Shutdown HTTP Server with timeout: %s", g.gracefulTimeout)

	// shut down gracefully, but wait no longer than timeout before halting
	ctx, cancel := context.WithTimeout(context.Background(), g.gracefulTimeout)
	defer cancel()
	g.httpServer.Shutdown(ctx)

	log.Info().Msg("HTTP Server gracefully stopped")
}
