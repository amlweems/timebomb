package engine

import (
	"fmt"
)

type State int

const (
	StateLobby State = iota
	StatePlaying
	StateDefenderWin
	StateBomberWin
)

func (s State) String() string {
	switch s {
	case StateLobby:
		return "lobby"
	case StatePlaying:
		return "playing"
	case StateDefenderWin:
		return "defenders win"
	case StateBomberWin:
		return "bombers win"
	default:
		return "unknown"
	}
}

type PlayerID int

type Card int

const (
	CardNop Card = iota
	CardWire
	CardBomb
)

func (c Card) String() string {
	switch c {
	case CardNop:
		return "-"
	case CardWire:
		return "+"
	case CardBomb:
		return "*"
	default:
		return "?"
	}
}

type Role int

const (
	RoleDefender Role = iota
	RoleBomber
)

func (r Role) String() string {
	switch r {
	case RoleDefender:
		return "defender"
	case RoleBomber:
		return "bomber"
	default:
		return "unknown"
	}
}

type Player struct {
	Name  string
	Cards []Card
	Role  Role
}

type Cut struct {
	Source PlayerID
	Target PlayerID
	Card   Card
}

type Game struct {
	Code string

	Players []Player

	State   State
	Round   int
	Nippers PlayerID
	Cuts    []Cut

	Wires int
	Bomb  bool

	subscribers []func(any)
}

type Event struct {
	Type string
}

func (g *Game) Subscribe(f func(any)) {
	g.subscribers = append(g.subscribers, f)
}

func (g *Game) Broadcast(x any) {
	for _, f := range g.subscribers {
		go f(x)
	}
}

func (g *Game) Join(name string) (PlayerID, error) {
	defer g.Broadcast(Event{Type: "join"})

	for id, player := range g.Players {
		if player.Name == name {
			return PlayerID(id), nil
		}
	}

	id := PlayerID(len(g.Players))
	if g.State == StatePlaying {
		return id, fmt.Errorf("cannot join game, already started")
	}
	g.Players = append(g.Players, Player{Name: name})
	return id, nil
}

func (g *Game) reset() {
	g.Cuts = []Cut{}
	g.Round = 0
	g.Wires = 0
	g.Bomb = false
}

func (g *Game) Start() error {
	defer g.Broadcast(Event{Type: "start"})

	if g.State == StatePlaying {
		return fmt.Errorf("cannot start game, already started")
	}
	n := len(g.Players)
	if n < 3 || n > 8 {
		return fmt.Errorf("invalid number of players: %d", n)
	}
	g.reset()
	g.assign()
	g.deal()
	g.Nippers = PlayerID(randN(len(g.Players)))
	g.State = StatePlaying
	return nil
}

func (g *Game) Cut(src, dst PlayerID, i int) error {
	defer g.Broadcast(Event{Type: "cut"})

	if g.State != StatePlaying {
		return fmt.Errorf("cannot cut card, game not started")
	}

	n := len(g.Players)
	if int(src) > n {
		return fmt.Errorf("invalid source player id: %d", src)
	}
	if int(dst) > n {
		return fmt.Errorf("invalid target player id: %d", dst)
	}
	if src == dst {
		return fmt.Errorf("cannot cut your own card")
	}
	if src != g.Nippers {
		return fmt.Errorf("not your turn")
	}

	// pick the target player's i-th card and remove it from their hand
	hand := g.Players[dst].Cards
	if i >= len(hand) {
		return fmt.Errorf("player only has %d cards", len(hand))
	}
	card := hand[i]
	g.Players[dst].Cards = append(hand[:i], hand[i+1:]...)

	// add the action to the log
	g.Cuts = append(g.Cuts, Cut{
		Source: src,
		Target: dst,
		Card:   card,
	})

	// pass the nippers to the target player
	g.Nippers = dst

	// update the card counts
	switch card {
	case CardBomb:
		g.Bomb = true
	case CardWire:
		g.Wires++
	case CardNop:
		// nothing happens
	}

	// if the game is over, assign the winner
	if g.Wires == n {
		g.State = StateDefenderWin
	} else if g.Bomb || g.Round == 4 {
		g.State = StateBomberWin
	}

	// if we've reached the end of the round, deal a new hand
	cuts := len(g.Cuts)
	if g.State == StatePlaying && cuts > 0 && cuts%n == 0 {
		g.deal()
		g.Round++
	}

	return nil
}
