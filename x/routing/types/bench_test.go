package types

import "testing"

func BenchmarkRouteFinancial(b *testing.B) {
	input := validFinancialInput()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		input.RoutingEpoch = uint64(i % 1024)
		decision, err := Route(input)
		if err != nil {
			b.Fatal(err)
		}
		if decision.ZoneID != ZoneFinancial {
			b.Fatalf("unexpected zone %s", decision.ZoneID)
		}
	}
}

func BenchmarkSortDecisions(b *testing.B) {
	decisions := make([]RouteDecision, 0, 512)
	for i := 0; i < 512; i++ {
		input := validFinancialInput()
		input.FeeClass = FeeClass(i % int(MaxFeeClass+1))
		input.ReputationClass = ReputationClass((512 - i) % int(MaxReputationClass+1))
		input.AdmissionHeight = uint64(i % 64)
		input.TxHash = hashBytes(string(rune('a' + (i % 26))))
		decision, err := Route(input)
		if err != nil {
			b.Fatal(err)
		}
		decisions = append(decisions, decision)
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ordered := SortDecisions(decisions)
		if len(ordered) != len(decisions) {
			b.Fatal("decision count changed")
		}
	}
}
