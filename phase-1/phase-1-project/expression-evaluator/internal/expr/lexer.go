package expr

// input statement: "2 + 3 * 4"
// output statement: Number("2"), Plus("+"), Number("3"), Star("*") and Number("4")


var keywords = map[string]TokenKind {
	"let": LET,
	"true": TRUE,
	"false": FALSE,
}
var singleOps = map[byte]TokenKind{
	'+': PLUS,
	'-': MINUS,
	'*': STAR,
	'/': SLASH,
	'%': PERCENT,
	'^': CARET,
	'!': BANG,
	'=': ASSIGN,
	'<': LT,
	'>': GT,
	'(': LPAREN,
	')': RPAREN,
	',': COMMA,
	'?': QUESTION,
	':': COLON,
}
var doubleOps = map[string]TokenKind{
	"==": EQ,
	"!=": NEQ,
	"<=": LTE,
	">=": GTE,
	"&&": AND,
	"||": OR,
}

type lexer struct {
	input	string
	pos		int
}

func New(input string) *lexer {
	return &lexer{
		input: input,
	}
}

// skip any whitespace inside the expression by removing ' ', '\t', '\r', '\n'
func (l *lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		
		// when the function encounters the unwanted expressions, it increments the position by 1
		if (ch != ' ') && (ch != '\t') && (ch != '\r') && (ch != '\n') {
			return
		}
		l.pos++
	}
}

// function to check if the character is a digit
func (l *lexer) isDigit(ch byte) bool {
	if ch >= '0' && ch <= '9' {
		return true
	}
	return false
}

// function to check if the character is a letter or '_' i.e. an identifier
func (l *lexer) isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// input statement: "123.456 + 2"
// Number tokens: 123.456, 2
// Plus token: +
// remaining whitespaces are removed
func (l *lexer) readNumber() Token {
	// remembering the start position for slicing the text
	startPos := l.pos

	// for the digits before the decimal point
	for (l.pos < len(l.input) && l.isDigit(l.input[l.pos])) {
		l.pos++
	}

	// for the digits after the decimal point
	if (l.pos < len(l.input) && l.input[l.pos] == '.') {
		l.pos++

		for (l.pos < len(l.input) && l.isDigit(l.input[l.pos])) {
			l.pos++
		}
	}

	return Token{
		Kind: NUMBER,
		Text: l.input[startPos:l.pos],
		Pos: startPos,
	}
	
}

// identifiers -> starts with a letter or _
// example: let x = 10
// let is saved as Token{LET, "let", 0}
// then whitespace is skipped and x is saved as Token{IDENT, "x", 4}
// then no more identifiers, different function called
func (l *lexer) readIdentifier() Token {
	startPos := l.pos

	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		// if the character at that position is not a letter, digit or underscore, break from the for loop
		// else advance the character position
		if !l.isDigit(ch) || !l.isLetter(ch) || ch != '_' {
			break
		}
		l.pos++
	}

	// checking whether any keyword matches or not with the sliced text
	text := l.input[startPos:l.pos]

	// if the sliced text matches with the keywords present in the map, then save it as a keyword
	// else save it as IDENT
	if kind, ok := keywords[text]; ok {
		return Token{
			Kind: kind,
			Text: text,
			Pos: startPos,
		}
	} else {
		return Token{
			Kind: IDENT,
			Text: text,
			Pos: startPos,
		}
	}
}

// testing the operators, both the single-character and double-character ones
func (l *lexer) readOperator() Token {
	startPos := l.pos

	// first testing the double-character operators so that the single-character operators don't get stuck
	if l.pos+1 < len(l.input) {
		text := l.input[l.pos:l.pos+2]
		if kind, ok := doubleOps[text]; ok {
			l.pos+=2
			return Token{Kind: kind, Text: text, Pos: startPos}
		}
	}
	
	ch := l.input[l.pos]
	l.pos++
	if kind, ok := singleOps[ch]; ok {
		return Token{Kind: kind, Text: string(ch), Pos: startPos}
	}

	return Token{
		Kind: ILLEGAL,
		Text: string(ch),
		Pos: startPos,
	}
}

// example: "Hello World"
// this function runs when we encounter the opening "
// and closes with the closing "
func (l *lexer) readString() Token {
	startPos := l.pos
	l.pos++

	for (l.pos < len(l.input)) {
		// checking whether the current character is a closing "
		// if it is, then consume the text within ""
		// or if not, then return the token as ILLEGAL
		if l.input[l.pos] == '"' {
			// slice the text until the closing '"'
			text := l.input[startPos+1:l.pos] // contents without the quotes
			l.pos++
			return Token{Kind: STRING, Text: text, Pos: startPos}
		}

		// if the character is not the closing '"', then simply advance the position
		l.pos++
	}

	// return the token as ILLEGAL if the string is not resolved in the for loop
	// meaning there is no closing '"' or the position has gone out of the length of the input
	return Token{Kind: ILLEGAL, Pos: startPos}
}

func (l *lexer) Next() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return Token{
			Kind: EOF,
			Pos: l.pos,
		}
	}

	ch := l.input[l.pos]

	// for digits
	if l.isDigit(ch) {
		 return l.readNumber()
	} 

	// for letters or _ called as identifiers
	if l.isLetter(ch) || ch == '_' {
		return l.readIdentifier()
	}

	// for strings
	if ch == '"' {
		return l.readString()
	}

	// operators and punctuation
	return l.readOperator()
}