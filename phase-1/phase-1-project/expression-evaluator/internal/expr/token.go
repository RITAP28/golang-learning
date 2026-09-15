package expr

// example statement: 2+3*4
// breaking down the above statement: Number, Plus, Star, EOF and more

type TokenKind int

// associating an integer with each token type
const (
	EOF TokenKind = iota	// 0
	ILLEGAL					// 1

	NUMBER					// 2
	STRING					// 3
	IDENT					// 4
	TRUE					// 5
	FALSE					// 6
	LET						// 7
	
	PLUS					// 8 for addition +
	MINUS					// 9 for substraction -
	STAR					// 10 for multiplication *
	SLASH					// 11 for division /
	PERCENT					// 12 for %
	CARET					// 13 for ^
	BANG					// 14 for !

	EQ
	NEQ
	LT
	LTE
	GT
	GTE
	AND
	OR

	ASSIGN
	LPAREN
	RPAREN
	COMMA
	QUESTION
	COLON
)

type Token struct {
	Kind	TokenKind	// declared above
	Text	string		// the raw text on the statement, for error statements
	Pos		int			// byte offset in the input
}

func (k TokenKind) String() string {
	switch k {
		case EOF:
			return "EOF"
		case ILLEGAL:
			return "ILLEGAL"
			
		case NUMBER:
			return "NUMBER"
		case STRING:
			return "STRING"
		case IDENT:
			return "IDENT"
		case TRUE:
			return "TRUE"
		case FALSE:
			return "FALSE"
		case LET:
			return "LET"
			
		case PLUS:
			return "+"
		case MINUS:
			return "-"
		case STAR:
			return "*"
		case SLASH:
			return "/"
		case PERCENT:
			return "%"
		case CARET:
			return "^"
		case BANG:
			return "!"

		case EQ:
			return "=="
		case NEQ:
			return "!="
		case LT:
			return "<"
		case LTE:
			return "<="
		case GT:
			return ">"
		case GTE:
			return ">="
		
		case AND:
			return "&&"
		case OR:
			return "||"
		
		case ASSIGN:
			return "="
		case LPAREN:
			return "("
		case RPAREN:
			return ")"
		case COMMA:
			return ","
		case QUESTION:
			return "?"
		case COLON:
			return ":"
			
		default:
			return "UNKNOWN"
	}
}















