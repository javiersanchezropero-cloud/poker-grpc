package eval

import (
	"math/rand"
	"time"

	"poker-grpc/backend/internal/cards"
)

var allSuits = []string{"H", "D", "C", "S"}
var allRanks = []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}

func newDeck() []cards.Card {
	deck := make([]cards.Card, 0, 52)
	for _, s := range allSuits {
		for _, r := range allRanks {
			deck = append(deck, cards.Card{Rank: r, Suit: s})
		}
	}
	return deck
}

func sameCard(a, b cards.Card) bool {
	return a.Rank == b.Rank && a.Suit == b.Suit
}

func removeKnown(deck []cards.Card, known []cards.Card) []cards.Card {
	out := make([]cards.Card, 0, len(deck))
	for _, c := range deck {
		keep := true
		for _, k := range known {
			if sameCard(c, k) {
				keep = false
				break
			}
		}
		if keep {
			out = append(out, c)
		}
	}
	return out
}

// WinProbMonteCarlo simulates hero vs (numPlayers-1) opponents.
// communityKnown must have length 0/3/4/5.
func WinProbMonteCarlo(heroHole []cards.Card, communityKnown []cards.Card, numPlayers int, numSims int) (float64, float64) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	wins := 0
	ties := 0

	for sim := 0; sim < numSims; sim++ {

		deck := newDeck()

		known := append([]cards.Card{}, heroHole...)
		known = append(known, communityKnown...)
		deck = removeKnown(deck, known)

		rng.Shuffle(len(deck), func(i, j int) {
			deck[i], deck[j] = deck[j], deck[i]
		})

		drawPos := 0

		// Complete community to 5
		community := append([]cards.Card{}, communityKnown...)
		for len(community) < 5 {
			community = append(community, deck[drawPos])
			drawPos++
		}

		// Deal opponents
		opponents := make([][]cards.Card, numPlayers-1)
		for i := 0; i < numPlayers-1; i++ {
			opponents[i] = []cards.Card{
				deck[drawPos],
				deck[drawPos+1],
			}
			drawPos += 2
		}

		// Evaluate hero
		hero7 := append([]cards.Card{}, heroHole...)
		hero7 = append(hero7, community...)
		_, heroRank := BestOf7(hero7)

		heroWins := true
		heroTies := false

		for _, opp := range opponents {
			opp7 := append([]cards.Card{}, opp...)
			opp7 = append(opp7, community...)
			_, oppRank := BestOf7(opp7)

			cmp := Compare(heroRank, oppRank)

			if cmp < 0 {
				heroWins = false
				heroTies = false
				break
			}

			if cmp == 0 {
				heroTies = true
			}
		}

		if heroWins {
			if heroTies {
				ties++
			} else {
				wins++
			}
		}
	}

	return float64(wins) / float64(numSims), float64(ties) / float64(numSims)
}
