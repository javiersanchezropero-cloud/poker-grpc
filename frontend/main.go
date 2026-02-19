package main

import (
	"context"
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pokerv1 "poker-grpc/backend/poker-grpc/gen/poker/v1"
)

// Static files
//
//go:embed static/index.html
var indexHTML []byte

type server struct {
	backendAddr string
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	backend := os.Getenv("BACKEND_ADDR")
	if backend == "" {
		// Your deployed gRPC backend on Cloud Run (host:443)
		backend = "poker-backend-1045727208254.europe-west1.run.app:443"
	}

	s := &server{backendAddr: backend}

	mux := http.NewServeMux()

	// Web UI
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(indexHTML)
	})

	// REST -> gRPC bridge
	mux.HandleFunc("/api/evaluate", s.handleEvaluate)
	mux.HandleFunc("/api/compare", s.handleCompare)
	mux.HandleFunc("/api/probability", s.handleProbability)

	addr := ":" + port
	log.Printf("Frontend server listening on %s (BACKEND_ADDR=%s)", addr, backend)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func (s *server) dial() (*grpc.ClientConn, error) {
	creds := credentials.NewTLS(&tls.Config{})
	return grpc.Dial(s.backendAddr, grpc.WithTransportCredentials(creds))
}

type card struct {
	Code string `json:"code"`
}

type evalReq struct {
	Hole      []string `json:"hole"`
	Community []string `json:"community"`
}

type compareReq struct {
	HoleA      []string `json:"holeA"`
	CommunityA []string `json:"communityA"`
	HoleB      []string `json:"holeB"`
	CommunityB []string `json:"communityB"`
}

type probReq struct {
	HeroHole   []string `json:"heroHole"`
	Community  []string `json:"community"`
	NumPlayers int32    `json:"numPlayers"`
	NumSims    int32    `json:"numSims"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}

func toCards(codes []string) []*pokerv1.Card {
	out := make([]*pokerv1.Card, 0, len(codes))
	for _, c := range codes {
		out = append(out, &pokerv1.Card{Code: c})
	}
	return out
}

func (s *server) handleEvaluate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, 405, map[string]any{"error": "POST required"})
		return
	}

	var in evalReq
	if err := readJSON(r, &in); err != nil {
		writeJSON(w, 400, map[string]any{"error": "invalid JSON"})
		return
	}

	conn, err := s.dial()
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": "failed to dial backend"})
		return
	}
	defer conn.Close()

	client := pokerv1.NewPokerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.EvaluateBestHand(ctx, &pokerv1.EvaluateRequest{
		Hole:      toCards(in.Hole),
		Community: toCards(in.Community),
	})
	if err != nil {
		writeJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}

	writeJSON(w, 200, resp)
}

func (s *server) handleCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, 405, map[string]any{"error": "POST required"})
		return
	}

	var in compareReq
	if err := readJSON(r, &in); err != nil {
		writeJSON(w, 400, map[string]any{"error": "invalid JSON"})
		return
	}

	conn, err := s.dial()
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": "failed to dial backend"})
		return
	}
	defer conn.Close()

	client := pokerv1.NewPokerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.CompareHands(ctx, &pokerv1.CompareRequest{
		HoleA:      toCards(in.HoleA),
		CommunityA: toCards(in.CommunityA),
		HoleB:      toCards(in.HoleB),
		CommunityB: toCards(in.CommunityB),
	})
	if err != nil {
		writeJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}

	writeJSON(w, 200, resp)
}

func (s *server) handleProbability(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, 405, map[string]any{"error": "POST required"})
		return
	}

	var in probReq
	if err := readJSON(r, &in); err != nil {
		writeJSON(w, 400, map[string]any{"error": "invalid JSON"})
		return
	}

	conn, err := s.dial()
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": "failed to dial backend"})
		return
	}
	defer conn.Close()

	client := pokerv1.NewPokerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	resp, err := client.WinProbabilityMonteCarlo(ctx, &pokerv1.ProbabilityRequest{
		HeroHole:       toCards(in.HeroHole),
		Community:      toCards(in.Community),
		NumPlayers:     in.NumPlayers,
		NumSimulations: in.NumSims,
	})
	if err != nil {
		writeJSON(w, 400, map[string]any{"error": err.Error()})
		return
	}

	writeJSON(w, 200, resp)
}
