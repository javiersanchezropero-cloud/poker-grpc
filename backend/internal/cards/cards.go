package cards

import (
	"errors"
	"fmt"
	"strings"
)

type Card struct {
	Rank int    // 2-14 (A=14)
	Suit string // H, D, C, S
}

var rankMap = map[string]int{
	"2": 2, "3": 3, "4": 4, "5": 5, "6": 6,
	"7": 7, "8": 8, "9": 9,
	"T": 10,
	"J": 11,
	"Q": 12,
	"K": 13,
	"A": 14,
}

var validSuits = map[string]bool{
	"H": true,
	"D": true,
	"C": true,
	"S": true,
}

// Expected format: Suit + Rank, e.g. "HA", "S7", "CT"
func ParseCard(code string) (Card, error) {
	code = strings.ToUpper(strings.TrimSpace(code))

	if len(code) != 2 {
		return Card{}, errors.New("card must have exactly 2 characters")
	}

	suitStr := string(code[0])
	rankStr := string(code[1])

	if !validSuits[suitStr] {
		return Card{}, fmt.Errorf("invalid suit: %s", suitStr)
	}

	rank, ok := rankMap[rankStr]
	if !ok {
		return Card{}, fmt.Errorf("invalid rank: %s", rankStr)
	}

	return Card{
		Rank: rank,
		Suit: suitStr,
	}, nil
}

func ParseCards(codes []string) ([]Card, error) {
	seen := make(map[string]bool)
	result := make([]Card, 0, len(codes))

	for _, code := range codes {
		code = strings.ToUpper(strings.TrimSpace(code))

		if seen[code] {
			return nil, fmt.Errorf("duplicate card detected: %s", code)
		}
		seen[code] = true

		card, err := ParseCard(code)
		if err != nil {
			return nil, err
		}

		result = append(result, card)
	}

	return result, nil
}
