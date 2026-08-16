package gatepasses

import "testing"

func boolPtr(v bool) *bool {
	return &v
}

func TestResolveReturnabilitySecurityRules(t *testing.T) {
	tests := []struct {
		name       string
		policy     ReturnabilityPolicy
		requested  *bool
		defaultVal bool
		want       bool
		wantErr    bool
	}{
		{
			name:       "required rejects explicit false",
			policy:     ReturnabilityRequired,
			requested:  boolPtr(false),
			defaultVal: false,
			wantErr:    true,
		},
		{
			name:       "required explicit true",
			policy:     ReturnabilityRequired,
			requested:  boolPtr(true),
			defaultVal: false,
			want:       true,
		},
		{
			name:       "not allowed rejects true",
			policy:     ReturnabilityNotAllowed,
			requested:  boolPtr(true),
			defaultVal: false,
			wantErr:    true,
		},
		{
			name:       "not allowed remains false",
			policy:     ReturnabilityNotAllowed,
			requested:  boolPtr(false),
			defaultVal: true,
			want:       false,
		},
		{
			name:       "optional follows request",
			policy:     ReturnabilityOptional,
			requested:  boolPtr(true),
			defaultVal: false,
			want:       true,
		},
		{
			name:       "optional false request",
			policy:     ReturnabilityOptional,
			requested:  boolPtr(false),
			defaultVal: true,
			want:       false,
		},
		{
			name:       "optional uses default when request omitted",
			policy:     ReturnabilityOptional,
			requested:  nil,
			defaultVal: true,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveReturnability(
				tt.policy,
				tt.requested,
				tt.defaultVal,
			)

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr = %v", err, tt.wantErr)
			}

			if !tt.wantErr && got != tt.want {
				t.Fatalf("got = %v, want = %v", got, tt.want)
			}
		})
	}
}
