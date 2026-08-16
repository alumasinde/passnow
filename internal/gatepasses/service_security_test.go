package gatepasses

import (
	"testing"

	"gatepass/internal/approvals"
)

func TestResolveReturnability(t *testing.T) {
	cases := []struct {
		name       string
		policy     string
		requested  bool
		defaultVal bool
		want       bool
		wantErr    bool
	}{
		{"required always enables", ReturnabilityRequired, false, false, true, false},
		{"required explicit true", ReturnabilityRequired, true, false, true, false},
		{"not allowed rejects true", ReturnabilityNotAllowed, true, false, false, true},
		{"not allowed false", ReturnabilityNotAllowed, false, true, false, false},
		{"optional follows request", ReturnabilityOptional, true, false, true, false},
		{"optional follows default", ReturnabilityOptional, false, true, true, false},
		{"optional false", ReturnabilityOptional, false, false, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ResolveReturnability(tc.policy, tc.requested, tc.defaultVal)
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v, wantErr=%v", err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestApprovalTypesRemainExplicit(t *testing.T) {
	if approvals.ApproverRole == approvals.ApproverSpecificUser {
		t.Fatal("approval actor types must remain distinct")
	}
}
