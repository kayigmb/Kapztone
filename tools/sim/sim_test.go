package main

import (
	"math/rand/v2"
	"reflect"
	"testing"
)

func TestPlanIsTheSameForTheSameSeed(t *testing.T) {
	draw := func(seed uint64) ChainPlan {
		p, err := planChain(rand.New(rand.NewPCG(seed, 0)), [2]float64{29, -2}, 20, 100, 5, true)
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	if !reflect.DeepEqual(draw(42), draw(42)) {
		t.Fatal("same seed, different plan")
	}
	if reflect.DeepEqual(draw(42), draw(43)) {
		t.Fatal("different seeds, same plan")
	}
}

func TestPlanCarriesFivePerTypePerChain(t *testing.T) {
	count := func(transforms bool) map[string]int {
		p, err := planChain(rand.New(rand.NewPCG(1, 0)), [2]float64{29, -2}, 20, 100, 5, transforms)
		if err != nil {
			t.Fatal(err)
		}
		n := map[string]int{F3: p.F3, F4: p.F4Capacity}
		for _, l := range p.Lots {
			if l.Fraud != "" && l.Attempt != "" {
				t.Fatal("a lot carries two frauds")
			}
			n[l.Fraud]++
			n[l.Attempt]++
		}
		return n
	}
	for _, transforms := range []bool{true, false} {
		n := count(transforms)
		for _, k := range []string{F1, F2, F3, F4, F5, F6} {
			if n[k] != 5 {
				t.Errorf("transforms=%v: %s has %d cases, want 5", transforms, k, n[k])
			}
		}
	}
	if _, err := planChain(rand.New(rand.NewPCG(1, 0)), [2]float64{29, -2}, 20, 10, 5, true); err == nil {
		t.Error("10 lots can't carry 30 frauds, but the plan didn't say so")
	}
}

func TestOutcome(t *testing.T) {
	cases := []struct {
		fraud     bool
		mechanism string
		status    int
		verdict   string
		want      string
	}{
		{true, viaAPI, 409, "", prevented},
		{true, viaAPI, 200, "", missed},
		{true, viaAPI, 500, "", missed},
		{true, viaDB, 200, "failed", detected},
		{true, viaDB, 200, "warning", detected},
		{true, viaDB, 200, "verified", missed},
		{false, "none", 200, "verified", clean},
		{false, "none", 200, "warning", falseAlarm},
	}
	for _, c := range cases {
		if got := outcome(c.fraud, c.mechanism, c.status, c.verdict); got != c.want {
			t.Errorf("outcome(%v, %s, %d, %q) = %s, want %s", c.fraud, c.mechanism, c.status, c.verdict, got, c.want)
		}
	}
}
