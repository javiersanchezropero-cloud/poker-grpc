package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"google.golang.org/grpc"

	pokerv1 "poker-grpc/backend/poker-grpc/gen/poker/v1"
)

func main() {
	xlsxPath := "tests/Texas HoldEm Hand comparison test cases.xlsx"
	addr := "127.0.0.1:8080"

	f, err := excelize.OpenFile(xlsxPath)
	if err != nil {
		log.Fatalf("failed to open xlsx %q: %v", xlsxPath, err)
	}
	defer func() { _ = f.Close() }()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		log.Fatalf("no sheets in %q", xlsxPath)
	}
	sheet := sheets[0]

	rows, err := f.GetRows(sheet)
	if err != nil {
		log.Fatalf("failed to read sheet %q: %v", sheet, err)
	}
	if len(rows) < 2 {
		log.Fatalf("sheet %q has no data rows", sheet)
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial %s: %v", addr, err)
	}
	defer conn.Close()

	client := pokerv1.NewPokerServiceClient(conn)

	total := 0
	passed := 0
	skipped := 0

	for i := 1; i < len(rows); i++ {
		rowNum := i + 1

		if len(rows[i]) < 7 {
			skipped++
			continue
		}

		communityStr := strings.TrimSpace(rows[i][1])
		p1Str := strings.TrimSpace(rows[i][2])
		p2Str := strings.TrimSpace(rows[i][4])
		resultStr := strings.TrimSpace(rows[i][6])

		if communityStr == "" || p1Str == "" || p2Str == "" || resultStr == "" {
			skipped++
			continue
		}

		community := splitCards(communityStr)
		p1 := splitCards(p1Str)
		p2 := splitCards(p2Str)

		if len(community) != 5 || len(p1) != 2 || len(p2) != 2 {
			skipped++
			continue
		}

		// Skip rows with duplicate cards across all inputs
		allCards := append([]string{}, community...)
		allCards = append(allCards, p1...)
		allCards = append(allCards, p2...)

		if hasDuplicates(allCards) {
			skipped++
			continue
		}

		exp, ok := parseExpected(resultStr)
		if !ok {
			skipped++
			continue
		}

		total++

		req := &pokerv1.CompareRequest{
			HoleA: []*pokerv1.Card{
				{Code: p1[0]}, {Code: p1[1]},
			},
			CommunityA: []*pokerv1.Card{
				{Code: community[0]}, {Code: community[1]},
				{Code: community[2]}, {Code: community[3]},
				{Code: community[4]},
			},
			HoleB: []*pokerv1.Card{
				{Code: p2[0]}, {Code: p2[1]},
			},
			CommunityB: []*pokerv1.Card{
				{Code: community[0]}, {Code: community[1]},
				{Code: community[2]}, {Code: community[3]},
				{Code: community[4]},
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		resp, err := client.CompareHands(ctx, req)
		cancel()

		if err != nil {
			fmt.Printf("FAIL row %d: RPC error: %v\n", rowNum, err)
			continue
		}

		got := resp.Result
		if got == exp {
			passed++
		} else {
			fmt.Printf("FAIL row %d: expected %v, got %v | community=%v p1=%v p2=%v\n",
				rowNum, exp, got, community, p1, p2)
		}
	}

	fmt.Println("---- SUMMARY ----")
	fmt.Printf("Total executed: %d\n", total)
	fmt.Printf("Passed: %d\n", passed)
	fmt.Printf("Failed: %d\n", total-passed)
	fmt.Printf("Skipped: %d\n", skipped)
}

func splitCards(s string) []string {
	parts := strings.Fields(s)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func hasDuplicates(codes []string) bool {
	seen := make(map[string]bool)
	for _, c := range codes {
		if seen[c] {
			return true
		}
		seen[c] = true
	}
	return false
}

func parseExpected(s string) (pokerv1.CompareResult, bool) {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "hand 1 > hand 2":
		return pokerv1.CompareResult_A_WINS, true
	case "hand 1 < hand 2":
		return pokerv1.CompareResult_B_WINS, true
	case "hand 1 = hand 2":
		return pokerv1.CompareResult_TIE, true
	default:
		return pokerv1.CompareResult_COMPARE_RESULT_UNSPECIFIED, false
	}
}
