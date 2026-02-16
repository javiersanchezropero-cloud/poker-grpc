package main

import (
	"context"
	"log"
	"time"

	pokerv1 "poker-grpc/backend/poker-grpc/gen/poker/v1"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pokerv1.NewPokerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// A: straight (2-6)
	req := &pokerv1.CompareRequest{
		HoleA:      []*pokerv1.Card{{Code: "HA"}, {Code: "HK"}},
		CommunityA: []*pokerv1.Card{{Code: "S2"}, {Code: "D3"}, {Code: "C4"}, {Code: "H5"}, {Code: "S6"}},

		// B: high card only
		HoleB:      []*pokerv1.Card{{Code: "D9"}, {Code: "C8"}},
		CommunityB: []*pokerv1.Card{{Code: "S2"}, {Code: "D3"}, {Code: "C4"}, {Code: "H9"}, {Code: "SK"}},
	}

	resp, err := client.CompareHands(ctx, req)
	if err != nil {
		log.Printf("RPC error: %v", err)
		return
	}

	log.Printf("Result: %v", resp.Result)
	log.Printf("A best: %+v", resp.BestA)
	log.Printf("B best: %+v", resp.BestB)
}
