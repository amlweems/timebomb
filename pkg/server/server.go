package server

import (
	"fmt"
	"time"

	"github.com/amlweems/timebomb/pkg/engine"
	"github.com/amlweems/timebomb/pkg/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
)

type Server struct {
	Games map[string]*engine.Game
}

func NewServer() *Server {
	return &Server{
		Games: make(map[string]*engine.Game),
	}
}

func (s *Server) Handler(r ssh.Session) (tea.Model, []tea.ProgramOption) {
	_, _, active := r.Pty()
	if !active {
		fmt.Fprintf(r, "error: no tty allocated, try `ssh -t`")
		return nil, nil
	}

	code := ticket()
	if len(r.Command()) > 0 {
		code = r.Command()[0]
	}

	game, ok := s.Games[code]
	if !ok {
		game = &engine.Game{}
		s.Games[code] = game
		time.AfterFunc(24*time.Hour, func() {
			delete(s.Games, code)
		})
	}
	player, err := game.Join(r.User())
	if err != nil {
		fmt.Fprintf(r, "error: %s\n", err)
		return nil, nil
	}

	m := tui.Model{
		Game:   game,
		Code:   code,
		Player: player,
	}
	return m, []tea.ProgramOption{
		tea.WithAltScreen(),
		func(p *tea.Program) {
			game.Subscribe(func(x any) {
				p.Send(x)
			})
		}}
}
