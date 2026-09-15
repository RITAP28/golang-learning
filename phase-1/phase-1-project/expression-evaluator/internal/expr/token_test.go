package expr

import "testing"

func TestTokenKindString(t *testing.T) {
	tests := []struct {
		kind 	TokenKind
		want	string
	}{
		{PLUS, "+"},
		{MINUS, "-"},
		{STAR, "*"},
		{NUMBER, "NUMBER"},
		{EOF, "EOF"},
	}

	for _, tt := range tests {
		got := tt.kind.String()

		if got != tt.want {
			t.Errorf("got %q, want %q", got, tt.want)
		}
	}
}