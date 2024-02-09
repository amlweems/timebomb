package engine

func (g *Game) assign() {
	roles := shuffle(rolesMap[len(g.Players)])
	for i := range g.Players {
		g.Players[i].Role = roles[i]
	}
}

func (g *Game) deal() {
	n := len(g.Players)
	cards := g.deck()
	size := len(cards) / n
	for i := range g.Players {
		g.Players[i].Cards = cards[i*size : i*size+size]
	}
}

func (g *Game) deck() []Card {
	n := len(g.Players)
	cards := make([]Card, 5*n-len(g.Cuts))
	if len(cards) == 0 {
		return cards
	}
	// 1 bomb
	cards[0] = CardBomb
	// n wires
	for i := 1; i < n+1-g.Wires; i++ {
		cards[i] = CardWire
	}
	// n*4-1 nops
	for i := n + 1; i < len(cards); i++ {
		cards[i] = CardNop
	}
	return shuffle(cards)
}

func (g *Game) RolesCount() (int, int) {
	roles := rolesMap[len(g.Players)]

	defenders := 0
	bombers := 0

	for _, role := range roles {
		switch role {
			case RoleDefender:
				defenders++
			case RoleBomber:
				bombers++
		}
	}

	return defenders, bombers
}

var rolesMap = map[int][]Role{
	3: {
		RoleDefender, RoleDefender,
		RoleBomber, RoleBomber,
	},
	4: {
		RoleDefender, RoleDefender, RoleDefender,
		RoleBomber, RoleBomber,
	},
	5: {
		RoleDefender, RoleDefender, RoleDefender,
		RoleBomber, RoleBomber,
	},
	6: {
		RoleDefender, RoleDefender, RoleDefender, RoleDefender,
		RoleBomber, RoleBomber,
	},
	7: {
		RoleDefender, RoleDefender, RoleDefender, RoleDefender, RoleDefender,
		RoleBomber, RoleBomber, RoleBomber,
	},
	8: {
		RoleDefender, RoleDefender, RoleDefender, RoleDefender, RoleDefender,
		RoleBomber, RoleBomber, RoleBomber,
	},
}
