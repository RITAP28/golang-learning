package expr

import "testing"

// testing basic arithmetic input statements
func TestLexerBasicArithemtic(t *testing.T) {
	tests := []struct{
		input	string
		want	[]TokenKind
	}{
		{
			input: "2 + 3 * 4",
			want: []TokenKind{
				NUMBER,
				PLUS,
				NUMBER,
				STAR,
				NUMBER,
				EOF,
			},
		},
		{
			input: "10 - 5 / 2",
			want: []TokenKind{
				NUMBER,
				MINUS,
				NUMBER,
				SLASH,
				NUMBER,
				EOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lex := New(tt.input)

			for i, wantKind := range tt.want {
				got := lex.Next()

				if got.Kind != wantKind {
					t.Fatalf(
						"token %d: got %s, want %s",
						i,
						got.Kind,
						wantKind,
					)
				}
			}
		})
	}
}

// testing numbers like integers and decimal numbers
// making sure readNumber() handles both integers and decimals
func TestLexerNumbers(t *testing.T) {
	tests := []struct{
		input	string
		text	string
	}{
		{"42", "42"},
		{"3.14", "3.14"},
		{"0", "0"},
		{"0.5", "0.5"},
		{"20.69", "20.69"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func (t *testing.T) {
			lex := New(tt.input)

			got := lex.Next()

			if got.Kind != NUMBER {
				t.Fatalf("got kind %s, want NUMBER", got.Kind)
			}

			if got.Text != tt.text {
				t.Fatalf("got text %q, want %q", got.Text, tt.text)
			}

			if got.Pos != 0 {
				t.Errorf("got position %d, want 0", got.Pos)
			}

			if eof := lex.Next(); eof.Kind != EOF {
				t.Errorf("got %s, want EOF", eof.Kind)
			}
		})
	}
}

// testing whitespace
// making sure input statements like "2+3", "2 + 3" and "\t2\n+\r3" give the same token sequence
func TestLexerWhitespace(t *testing.T) {
	tests := []struct{
		input	string
		want	[]TokenKind
	}{
		{
			input: "2+3",
			want: []TokenKind{
				NUMBER,
				PLUS,
				NUMBER,
				EOF,
			},
		},
		{
			input: "2 + 3",
			want: []TokenKind{
				NUMBER,
				PLUS,
				NUMBER,
				EOF,
			},
		},
		{
			input: "\t2\n+\r3",
			want: []TokenKind{
				NUMBER,
				PLUS,
				NUMBER,
				EOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func (t *testing.T) {
			lex := New(tt.input)

			for i, wantKind := range tt.want {
				got := lex.Next()

				if got.Kind != wantKind {
					t.Fatalf(
						"token %d: got %s, want %s",
						i,
						got.Kind,
						wantKind,
					)
				}
			}
		})
	}
}

// testing identifiers
func TestLexerIdentifiers(t *testing.T) {
	tests := []struct{
		input	string
		want	TokenKind
	}{
		{"x", IDENT},
		{"foo", IDENT},
		{"myVariable", IDENT},
		{"hello123", IDENT},
		{"user_123", IDENT},
		{"_", IDENT},
	}

	for _, tt := range tests {
		t.Run(tt.input, func (t *testing.T) {
			lex := New(tt.input)

			got := lex.Next()

			if got.Kind != tt.want {
				t.Fatalf("expected %s, got %s", tt.want, got.Kind)
			}

			if got.Text != tt.input {
				t.Errorf("got %q, want %q", got.Text, tt.input)
			}
		})
	}
}

// testing keywords, a subpart of identifiers like LET, TRUE and FALSE
func TestLexerKeywords(t *testing.T) {
	tests := []struct{
		input	string
		want	TokenKind
	}{
		{"let", LET},
		{"true", TRUE},
		{"false", FALSE},
		{"trueValue", IDENT},
		{"falseHood", IDENT},
	}

	for _, tt := range tests {
		t.Run(tt.input, func (t *testing.T) {
			lex := New(tt.input)
			got := lex.Next()

			if got.Kind != tt.want {
				t.Errorf("expected %s, got %s", tt.want, got.Kind)
			}
		})
	}
}

// testing operators like +, -, *, / etc
func TestLexerSingleCharacterOperators(t *testing.T) {
	tests := []struct{
		input	string
		want	TokenKind
	}{
		{"+", PLUS},
		{"-", MINUS},
		{"*", STAR},
		{"/", SLASH},
		{"%", PERCENT},
		{"^", CARET},
		{"!", BANG},
		{"<", LT},
		{">", GT},
		{"=", ASSIGN},
		{"(", LPAREN},
		{")", RPAREN},
		{",", COMMA},
		{"?", QUESTION},
		{":", COLON},
	}

	for _, tt := range tests {
		t.Run(tt.input, func (t *testing.T) {
			input := New(tt.input)
			result := input.Next()

			if result.Kind != tt.want {
				t.Errorf("expected %s, got %s", tt.want, result.Kind)
			}

			if eof := input.Next(); eof.Kind != EOF {
				t.Errorf("got %s, expected EOF", eof.Kind)
			}
		})
	}
}

// testing operators like ==, >=, <= etc
func TestLexerTwoCharacterOperators(t *testing.T) {
	tests := []struct{
		input	string
		want	TokenKind
	}{
		{"==", EQ},
		{"!=", NEQ},
		{"<=", LTE},
		{">=", GTE},
		{"&&", AND},
		{"||", OR},
	}

	for _, tt := range tests {
		t.Run(tt.input, func (t *testing.T) {
			lex := New(tt.input)
			got := lex.Next()

			if got.Kind != tt.want {
				t.Errorf("expected %s, got %s", tt.want, got.Kind)
			}

			if eof := lex.Next(); eof.Kind != EOF {
				t.Errorf("got %s, expected EOF", eof.Kind)
			}
		})
	}
}

// testing strings
func TestLexerStrings(t *testing.T) {
	tests := []struct{
		input	string
		want	TokenKind
	}{
		{`"hello"`, STRING},
		{`"hello world"`, STRING},
		{`"123"`, STRING},
		{`""`, STRING},
	}

	for _, tt := range tests {
		t.Run(tt.input, func (t *testing.T) {
			input := New(tt.input)
			result := input.Next()

			if result.Kind != tt.want {
				t.Errorf("expected %s, got %s", tt.want, result.Kind)
			}

			if eof := input.Next(); eof.Kind != EOF {
				t.Errorf("got %s, expected EOF", eof.Kind)
			}
		})
	}
}