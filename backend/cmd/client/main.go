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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	resp, err := client.EvaluateBestHand(ctx, &pokerv1.EvaluateRequest{
		Hole: []*pokerv1.Card{
			{Code: "HA"},
			{Code: "HK"},
		},
		Community: []*pokerv1.Card{
			{Code: "S2"},
			{Code: "D3"},
			{Code: "C4"},
			{Code: "H5"},
			{Code: "S6"},
		},
	})

	if err != nil {
		log.Printf("RPC error: %v", err)
		return
	}

	log.Printf("Response: %+v", resp)
}
