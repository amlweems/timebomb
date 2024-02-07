package tui

import (
	"fmt"
	"strings"

	"github.com/amlweems/timebomb/pkg/engine"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

var (
	color = termenv.EnvColorProfile().Color
	red = termenv.Style{}.Foreground(color("9")).Styled
	blue = termenv.Style{}.Foreground(color("12")).Styled
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

		sb.WriteString("\n")
		if nc > 0 && nc%n == 0 && m.Game.State == engine.StatePlaying {
			fmt.Fprintf(&sb, "Round %d over!\n", m.Game.Round)
		} else {
			fmt.Fprintf(&sb, "Round: %d\n", m.Game.Round+1)
		}
		for i := r*n; i < len(m.Game.Cuts); i++ {
			cut := m.Game.Cuts[i]
			fmt.Fprintf(&sb, "%s cut %s: %s\n",
				m.Game.Players[cut.Source].Name,
				m.Game.Players[cut.Target].Name,
				cut.Card)
		}
		cutsLeft := n-nc%n
		if nc > 0 && nc%n == 0 {
			if m.Game.State == engine.StatePlaying {
				fmt.Fprintf(&sb, "\nRound: %d\n", m.Game.Round+1)
			} else {
				cutsLeft = 0
			}
		}
		fmt.Fprintf(&sb, "[%d cuts left]\n", cutsLeft)

		sb.WriteString("\n")

		self := m.Game.Players[m.Player]
		numDefenders, numBombers := m.Game.RolesCount()
		role := ""
		switch self.Role {
			case engine.RoleDefender:
				role = blue(self.Role.String())
			case engine.RoleBomber:
				role = red(self.Role.String())
		}
		fmt.Fprintf(&sb, "Role: %s\n", role)
		fmt.Fprintf(&sb, "Possible roles: %d " + blue("defenders") + ", %d " + red("bombers") + "\n", numDefenders, numBombers)
		fmt.Fprintf(&sb, "Nippers: %s\n", m.Game.Players[m.Game.Nippers].Name)
		fmt.Fprintf(&sb, "Wires: %d/%d (%d left)\n", m.Game.Wires, n, n-m.Game.Wires)
		fmt.Fprintf(&sb, "\nPlayers:\n")

		var maxLength int
		for _, player := range m.Game.Players {
			if n := len(player.Name); n > maxLength {
				maxLength = n
			}
		}

		for id, player := range m.Game.Players {
			fmt.Fprintf(&sb, " %d. %-*s ", id+1, maxLength, player.Name)
			for i, card := range player.Cards {
				if id != int(m.Player) &&  m.Game.State == engine.StatePlaying {
					card = -1
				}
				cardStr := ""
				switch card {
					case engine.CardWire:
						cardStr = blue(card.String())
					case engine.CardBomb:
						cardStr = red(card.String())
					default:
						cardStr = card.String()
				}
				if i == m.cc && id == m.pc && m.Game.State == engine.StatePlaying {
					sb.WriteString("[" + cardStr + "]")
				} else {
					sb.WriteString(" " + cardStr+ " ")
				}
			}
			sb.WriteRune('\n')
		}

		if m.Game.State == engine.StateBomberWin {
			sb.WriteString(red("\nBombers win!\n"))
		}
		if m.Game.State == engine.StateDefenderWin {
			sb.WriteString(blue("\nDefenders win!\n"))
		}
	}
	if m.err != nil {
		fmt.Fprintf(&sb, "\n%s\n", m.err)
	}
	return sb.String()
}
