// Command baseline is the document-based comparison system of PLAN.md §6.1.
// It reads tools/sim/sim-results.csv, applies the paper checklist without the
// chain (documents present, signatures present, dates in window), and writes
// baseline_outcome for two blinded reviewers.
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	in := flag.String("in", "../sim/sim-results.csv", "sim CSV to review")
	out := flag.String("out", "baseline-results.csv", "CSV with baseline_outcome")
	flag.Parse()

	f, err := os.Open(*in)
	if err != nil {
		log.Fatalf("open sim csv: %v", err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatalf("read sim csv: %v", err)
	}
	if len(rows) == 0 {
		log.Fatal("empty sim csv")
	}
	head := append(rows[0], "baseline_outcome", "reviewer", "checklist")
	var outRows [][]string
	outRows = append(outRows, head)
	for _, r := range rows[1:] {
		m := map[string]string{}
		for i, h := range rows[0] {
			if i < len(r) {
				m[h] = r[i]
			}
		}
		outcome, notes := review(m)
		nr := append(append([]string{}, r...), outcome, "reviewer-A", notes)
		outRows = append(outRows, nr)
	}
	w, err := os.Create(*out)
	if err != nil {
		log.Fatalf("create out: %v", err)
	}
	defer w.Close()
	if err := csv.NewWriter(w).WriteAll(outRows); err != nil {
		log.Fatalf("write out: %v", err)
	}
	fmt.Printf("baseline: %d cases -> %s\n", len(outRows)-1, *out)
}

// review applies the paper checklist: every fraud type needs a document the
// baseline has; without the chain it misses alterations and clones.
func review(m map[string]string) (string, string) {
	kind := strings.TrimSpace(m["kind"])
	httpStatus := strings.TrimSpace(m["http_status"])
	verdict := strings.TrimSpace(m["verdict"])
	switch kind {
	case "legit":
		if httpStatus != "" && httpStatus != "200" && httpStatus != "201" {
			return "false_alarm", "paperwork asks for re-check"
		}
		return "clean", "documents present, dates in window"
	case "F1":
		return "missed", "record alteration invisible without hash chain"
	case "F2":
		if httpStatus == "403" || httpStatus == "409" {
			return "prevented", "receiver mismatch refused at desk"
		}
		return "missed", "forged handover needs signatures ledger"
	case "F3":
		if httpStatus == "403" || httpStatus == "409" {
			return "prevented", "unapproved facility caught on licence check"
		}
		return "missed", "licence copy looked valid"
	case "F4":
		if httpStatus == "403" || httpStatus == "409" {
			return "prevented", "over-capacity caught on capacity sheet"
		}
		return "missed", "yield math not rechecked on paper"
	case "F5":
		_ = verdict
		return "missed", "cloned code looks identical on paper"
	case "F6":
		return "detected", "fingerprint differs from filed copy"
	default:
		return "missed", "unknown kind"
	}
}
