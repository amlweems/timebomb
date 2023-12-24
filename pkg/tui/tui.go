package tui

import (
	"fmt"
	"strings"

	"github.com/amlweems/timebomb/pkg/engine"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Game   *engine.Game
	Code   string
	Player engine.PlayerID

	pc int
	cc int

	err error
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "s":
			m.err = m.Game.Start()
		case "c":
			m.err = m.Game.Cut(m.Player, engine.PlayerID(m.pc), m.cc)
		case "?":
			m.err = fmt.Errorf("press 's' to start, press 'c' to cut")
		case "down":
			m.pc++
		case "up":
			m.pc--
		case "right":
			m.cc++
		case "left":
			m.cc--
		case "q":
			return m, tea.Quit
		}
	}
	if m.pc < 0 {
		m.pc = 0
	}
	if n := len(m.Game.Players); m.pc >= n {
		m.pc = n - 1
	}
	if m.cc < 0 {
		m.cc = 0
	}
	if n := len(m.Game.Players[m.pc].Cards); m.cc >= n {
		m.cc = n - 1
	}

	return m, nil
}

func (m Model) View() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "State: %s\n", m.Game.State)
	switch m.Game.State {
	case engine.StateLobby:
		fmt.Fprintf(&sb, "\nCode: %s\nPlayers:\n", m.Code)
		for _, player := range m.Game.Players {
			fmt.Fprintf(&sb, " %s\t\n", player.Name)
		}
	case engine.StatePlaying, engine.StateBomberWin, engine.StateDefenderWin:
		n := len(m.Game.Players)
		nc := len(m.Game.Cuts)
		r := m.Game.Round
		if nc%n == 0 {
			r -= 1
		}
		if r < 0 {
			r = 0
		}

		fmt.Fprintf(&sb, "\nRound: %d\n", 1+r)
		for i := r*n; i < len(m.Game.Cuts); i++ {
			cut := m.Game.Cuts[i]
			fmt.Fprintf(&sb, "%s cut %s: %s\n",
				m.Game.Players[cut.Source].Name,
				m.Game.Players[cut.Target].Name,
				cut.Card)
		}
		if nc > 0 && nc%n == 0 {
			fmt.Fprintf(&sb, "Round over.\n")
		}
		sb.WriteString("\n")

		self := m.Game.Players[m.Player]
		fmt.Fprintf(&sb, "Role: %s\n", self.Role)
		fmt.Fprintf(&sb, "Nippers: %s\n", m.Game.Players[m.Game.Nippers].Name)
		fmt.Fprintf(&sb, "Wires: %d\n", m.Game.Wires)
		fmt.Fprintf(&sb, "Players:\n")

		var maxLength int
		for _, player := range m.Game.Players {
			if n := len(player.Name); n > maxLength {
				maxLength = n
			}
		}

		for id, player := range m.Game.Players {
			fmt.Fprintf(&sb, " %d. %-*s ", id+1, maxLength, player.Name)
			for i, card := range player.Cards {
				if id != int(m.Player) {
					card = -1
				}
				if i == m.cc && id == m.pc {
					sb.WriteString("[" + card.String() + "]")
				} else {
					sb.WriteString(" " + card.String() + " ")
				}
			}
			sb.WriteRune('\n')
		}

		if m.Game.State == engine.StateBomberWin {
			sb.WriteString("\nBombers win!\n")
		}
		if m.Game.State == engine.StateDefenderWin {
			sb.WriteString("\nDefenders win!\n")
		}
	}
	if m.err != nil {
		fmt.Fprintf(&sb, "\n%s\n", m.err)
	}
	return sb.String()
}
