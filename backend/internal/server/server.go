package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"poker-grpc/backend/internal/cards"
	"poker-grpc/backend/internal/eval"
	pokerv1 "poker-grpc/backend/poker-grpc/gen/poker/v1"
)

type PokerServer struct {
	pokerv1.UnimplementedPokerServiceServer
}

func NewPokerServer() *PokerServer {
	return &PokerServer{}
}

func (s *PokerServer) EvaluateBestHand(ctx context.Context, req *pokerv1.EvaluateRequest) (*pokerv1.EvaluateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	if len(req.Hole) != 2 {
		return nil, status.Error(codes.InvalidArgument, "hole must contain exactly 2 cards")
	}
	if len(req.Community) != 5 {
		return nil, status.Error(codes.InvalidArgument, "community must contain exactly 5 cards")
	}

	// Collect 7 codes preserving order (hole first, then community)
	allCodes := make([]string, 0, 7)
	for _, c := range req.Hole {
		allCodes = append(allCodes, c.GetCode())
	}
	for _, c := range req.Community {
		allCodes = append(allCodes, c.GetCode())
	}

	parsed, err := cards.ParseCards(allCodes)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid cards: %v", err)
	}

	bestIdx, bestRank := eval.BestOf7(parsed)

	// Build best 5 cards using bestIdx
	bestCards := make([]*pokerv1.Card, 0, 5)
	for _, i := range bestIdx {
		bestCards = append(bestCards, &pokerv1.Card{Code: allCodes[i]})
	}

	return &pokerv1.EvaluateResponse{
		Best: &pokerv1.BestHand{
			Cards:      bestCards,
			Category:   pokerv1.HandCategory(bestRank.Category),
			RankVector: bestRank.RankVector,
		},
	}, nil
}

func (s *PokerServer) CompareHands(ctx context.Context, req *pokerv1.CompareRequest) (*pokerv1.CompareResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	if len(req.HoleA) != 2 || len(req.CommunityA) != 5 {
		return nil, status.Error(codes.InvalidArgument, "player A must have hole=2 and community=5")
	}
	if len(req.HoleB) != 2 || len(req.CommunityB) != 5 {
		return nil, status.Error(codes.InvalidArgument, "player B must have hole=2 and community=5")
	}

	// Collect codes (A)
	codesA := make([]string, 0, 7)
	for _, c := range req.HoleA {
		codesA = append(codesA, c.GetCode())
	}
	for _, c := range req.CommunityA {
		codesA = append(codesA, c.GetCode())
	}
	parsedA, err := cards.ParseCards(codesA)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid cards for player A: %v", err)
	}

	// Collect codes (B)
	codesB := make([]string, 0, 7)
	for _, c := range req.HoleB {
		codesB = append(codesB, c.GetCode())
	}
	for _, c := range req.CommunityB {
		codesB = append(codesB, c.GetCode())
	}
	parsedB, err := cards.ParseCards(codesB)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid cards for player B: %v", err)
	}

	// Evaluate best hands
	bestIdxA, bestRankA := eval.BestOf7(parsedA)
	bestIdxB, bestRankB := eval.BestOf7(parsedB)

	bestCardsA := make([]*pokerv1.Card, 0, 5)
	for _, i := range bestIdxA {
		bestCardsA = append(bestCardsA, &pokerv1.Card{Code: codesA[i]})
	}

	bestCardsB := make([]*pokerv1.Card, 0, 5)
	for _, i := range bestIdxB {
		bestCardsB = append(bestCardsB, &pokerv1.Card{Code: codesB[i]})
	}

	// Compare
	cmp := eval.Compare(bestRankA, bestRankB)
	result := pokerv1.CompareResult_TIE
	if cmp == 1 {
		result = pokerv1.CompareResult_A_WINS
	} else if cmp == -1 {
		result = pokerv1.CompareResult_B_WINS
	}

	return &pokerv1.CompareResponse{
		BestA: &pokerv1.BestHand{
			Cards:      bestCardsA,
			Category:   pokerv1.HandCategory(bestRankA.Category),
			RankVector: bestRankA.RankVector,
		},
		BestB: &pokerv1.BestHand{
			Cards:      bestCardsB,
			Category:   pokerv1.HandCategory(bestRankB.Category),
			RankVector: bestRankB.RankVector,
		},
		Result: result,
	}, nil
}

func (s *PokerServer) WinProbabilityMonteCarlo(ctx context.Context, req *pokerv1.ProbabilityRequest) (*pokerv1.ProbabilityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	if len(req.HeroHole) != 2 {
		return nil, status.Error(codes.InvalidArgument, "hero_hole must contain exactly 2 cards")
	}

	nc := len(req.Community)
	if !(nc == 0 || nc == 3 || nc == 4 || nc == 5) {
		return nil, status.Error(codes.InvalidArgument, "community must have size 0, 3, 4, or 5")
	}

	if req.NumPlayers < 2 {
		return nil, status.Error(codes.InvalidArgument, "num_players must be >= 2")
	}
	if req.NumSimulations <= 0 {
		return nil, status.Error(codes.InvalidArgument, "num_simulations must be > 0")
	}

	heroCodes := []string{req.HeroHole[0].GetCode(), req.HeroHole[1].GetCode()}

	commCodes := make([]string, 0, nc)
	for _, c := range req.Community {
		commCodes = append(commCodes, c.GetCode())
	}

	// Validate duplicates across all known cards + parse them
	allCodes := append(append([]string{}, heroCodes...), commCodes...)
	parsedAll, err := cards.ParseCards(allCodes)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid cards: %v", err)
	}

	heroHole := parsedAll[:2]
	community := parsedAll[2:]

	win, tie := eval.WinProbMonteCarlo(heroHole, community, int(req.NumPlayers), int(req.NumSimulations))

	return &pokerv1.ProbabilityResponse{
		WinProbability: win,
		TieProbability: tie,
	}, nil
}
