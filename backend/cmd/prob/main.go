package main

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	pokerv1 "poker-grpc/backend/poker-grpc/gen/poker/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	conn, err := grpc.Dial("poker-backend-1045727208254.europe-west1.run.app:443", grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pokerv1.NewPokerServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Example: Hero has HA HK. No community known. 2 players. 10000 sims.
	req := &pokerv1.ProbabilityRequest{
		HeroHole: []*pokerv1.Card{
			{Code: "HA"},
			{Code: "HK"},
		},
		Community:      []*pokerv1.Card{}, // 0 known
		NumPlayers:     2,
		NumSimulations: 10000,
	}

	resp, err := client.WinProbabilityMonteCarlo(ctx, req)
	if err != nil {
		log.Printf("RPC error: %v", err)
		return
	}

	log.Printf("WinProbability: %.4f", resp.WinProbability)
	log.Printf("TieProbability: %.4f", resp.TieProbability)
}
