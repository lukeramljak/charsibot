package stats

import (
	"strings"
	"testing"
)

func TestParseModifyStatCommand(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantLogin       string
		wantColumn      string
		wantAmount      int64
		wantErr         bool
		wantErrContains string
	}{
		{
			name:       "parses addstat command",
			input:      "!addstat @foo strength 3",
			wantLogin:  "foo",
			wantColumn: "strength",
			wantAmount: 3,
			wantErr:    false,
		},
		{
			name:       "parses rmstat command",
			input:      "!rmstat @bar luck 2",
			wantLogin:  "bar",
			wantColumn: "luck",
			wantAmount: 2,
			wantErr:    false,
		},
		{
			name:            "errors on missing mention",
			input:           "!addstat strength 3",
			wantErr:         true,
			wantErrContains: "expected format",
		},
		{
			name:            "errors on missing stat and amount",
			input:           "!addstat @user",
			wantErr:         true,
			wantErrContains: "expected format",
		},
		{
			name:            "errors on missing amount",
			input:           "!addstat @user strength",
			wantErr:         true,
			wantErrContains: "expected format",
		},
		{
			name:            "errors on invalid number",
			input:           "!addstat @user strength abc",
			wantErr:         true,
			wantErrContains: "invalid number",
		},
		{
			name:       "handles mention with @ symbol",
			input:      "!addstat @username strength 5",
			wantLogin:  "username",
			wantColumn: "strength",
			wantAmount: 5,
			wantErr:    false,
		},
		{
			name:       "converts username to lowercase",
			input:      "!addstat @UserName strength 5",
			wantLogin:  "username",
			wantColumn: "strength",
			wantAmount: 5,
			wantErr:    false,
		},
		{
			name:       "handles extra whitespace",
			input:      "!addstat  @user   strength   5",
			wantLogin:  "user",
			wantColumn: "strength",
			wantAmount: 5,
			wantErr:    false,
		},
		{
			name:       "parses negative numbers",
			input:      "!addstat @user strength -3",
			wantLogin:  "user",
			wantColumn: "strength",
			wantAmount: -3,
			wantErr:    false,
		},
		{
			name:       "parses zero",
			input:      "!addstat @user strength 0",
			wantLogin:  "user",
			wantColumn: "strength",
			wantAmount: 0,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			login, column, amount, err := parseModifyStatCommand(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErrContains)
				} else if tt.wantErrContains != "" && !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.wantErrContains)) {
					t.Errorf("expected error containing %q, got %q", tt.wantErrContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if login != tt.wantLogin {
				t.Errorf("login = %q, want %q", login, tt.wantLogin)
			}
			if column != tt.wantColumn {
				t.Errorf("column = %q, want %q", column, tt.wantColumn)
			}
			if amount != tt.wantAmount {
				t.Errorf("amount = %d, want %d", amount, tt.wantAmount)
			}
		})
	}
}
