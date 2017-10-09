package https

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog/log"
)

type GracefulHTTPS interface{}

type gracefulHTTPS struct {
	httpServer      *http.Server
	gracefulTimeout time.Duration
	tls             struct {
		certFile string
		keyFile  string
	}
}

func New(httpServer *http.Server, gracefulTimeout time.Duration, certFile string, keyFile string) *gracefulHTTPS {
	gracefulHTTPS := &gracefulHTTPS{
		httpServer:      httpServer,
		gracefulTimeout: gracefulTimeout,
	}
	gracefulHTTPS.tls.certFile = certFile
	gracefulHTTPS.tls.keyFile = keyFile

	return gracefulHTTPS
}

func (g *gracefulHTTPS) Run() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	go func() {
		log.Info().Msgf("HTTPS Listen Address: %s", g.httpServer.Addr)
		log.Error().Err(g.httpServer.ListenAndServeTLS(g.tls.certFile, g.tls.keyFile)).Msg("HTTP Server error")
	}()

	<-stopChan // wait for SIGINT
	log.Info().Msgf("Shutdown HTTPS Server with timeout: %s", g.gracefulTimeout)

	// shut down gracefully, but wait no longer than timeout before halting
	ctx, cancel := context.WithTimeout(context.Background(), g.gracefulTimeout)
	defer cancel()
	g.httpServer.Shutdown(ctx)

	log.Info().Msg("HTTPS Server gracefully stopped")
}
