package tui

import (
	"fmt"
	"strings"

	"github.com/amlweems/timebomb/pkg/engine"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	red  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	blue = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
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

func (m Model) printLastCuts(sb *strings.Builder, max int) {
	totalCuts := len(m.Game.Cuts)
	numCuts := max
	if max > len(m.Game.Cuts) {
		numCuts = totalCuts
	}

	for i := totalCuts - numCuts; i < len(m.Game.Cuts); i++ {
		cut := m.Game.Cuts[i]
		fmt.Fprintf(sb, "%s cut %s: %s\n",
			m.Game.Players[cut.Source].Name,
			m.Game.Players[cut.Target].Name,
			cut.Card)
	}
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

		sb.WriteString("\n")
		if nc > 0 && nc%n == 0 {
			previousRound := m.Game.Round
			if m.Game.State != engine.StatePlaying {
				previousRound++
			}
			fmt.Fprintf(&sb, "Round %d over!\n", previousRound)
			m.printLastCuts(&sb, n)
		} else {
			fmt.Fprintf(&sb, "Round: %d\n", m.Game.Round+1)
			m.printLastCuts(&sb, nc%n)
		}
		cutsLeft := n - nc%n
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
			role = blue.Render(self.Role.String())
		case engine.RoleBomber:
			role = red.Render(self.Role.String())
		}
		fmt.Fprintf(&sb, "Role: %s\n", role)
		fmt.Fprintf(&sb, "Possible roles: %d "+blue.Render("defenders")+", %d "+red.Render("bombers")+"\n", numDefenders, numBombers)
		fmt.Fprintf(&sb, "Nippers: %s\n", m.Game.Players[m.Game.Nippers].Name)
		fmt.Fprintf(&sb, "Wires: %d/%d (%d left)\n", m.Game.Wires, n, n-m.Game.Wires)
		fmt.Fprintf(&sb, "\nPlayers:\n")

		var maxLength int
		var playerNames []string
		for _, player := range m.Game.Players {
			playerName := player.Name
			if m.Game.State != engine.StatePlaying {
				playerName = playerName + " (" + player.Role.String() + ")"
				switch player.Role {
				case engine.RoleDefender:
					playerName = blue.Render(playerName)
				case engine.RoleBomber:
					playerName = red.Render(playerName)
				}
			}
			if n := len(playerName); n > maxLength {
				maxLength = n
			}
			playerNames = append(playerNames, playerName)
		}

		for id, player := range m.Game.Players {
			fmt.Fprintf(&sb, " %d. %-*s ", id+1, maxLength, playerNames[id])
			for i, card := range player.Cards {
				if id != int(m.Player) && m.Game.State == engine.StatePlaying {
					card = -1
				}
				cardStr := ""
				switch card {
				case engine.CardWire:
					cardStr = blue.Render(card.String())
				case engine.CardBomb:
					cardStr = red.Render(card.String())
				default:
					cardStr = card.String()
				}
				if i == m.cc && id == m.pc && m.Game.State == engine.StatePlaying {
					sb.WriteString("[" + cardStr + "]")
				} else {
					sb.WriteString(" " + cardStr + " ")
				}
			}
			sb.WriteRune('\n')
		}

		if m.Game.State == engine.StateBomberWin {
			sb.WriteString(red.Render("\nBombers win!\n"))
		}
		if m.Game.State == engine.StateDefenderWin {
			sb.WriteString(blue.Render("\nDefenders win!\n"))
		}
	}
	if m.err != nil {
		fmt.Fprintf(&sb, "\n%s\n", m.err)
	}
	return sb.String()
}
