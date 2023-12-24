package main

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"

	"github.com/amlweems/timebomb/pkg/server"
)

//go:embed static
var static embed.FS

func main() {
	srv := server.NewServer()

	hostKeyOpt := wish.WithHostKeyPath(".ssh/term_info_ed25519")
	if key := os.Getenv("SSH_HOST_KEY"); key != "" {
		hostKeyOpt = wish.WithHostKeyPEM([]byte(key))
	}
	s, err := wish.NewServer(
		wish.WithAddress(":2222"),
		wish.WithMiddleware(
			bm.Middleware(srv.Handler),
			lm.Middleware(),
		),
		hostKeyOpt,
	)
	if err != nil {
		log.Fatal("could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server")
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("could not start ssh server", "error", err)
			done <- nil
		}
	}()

	index, _ := fs.Sub(static, "static")
	http.Handle("/", http.FileServer(http.FS(index)))
	go func() {
		if err := http.ListenAndServe(":8080", nil); !errors.Is(err, http.ErrServerClosed) {
			log.Error("could not start http server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("could not stop server", "error", err)
	}
}
