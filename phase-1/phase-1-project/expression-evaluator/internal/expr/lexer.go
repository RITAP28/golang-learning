package expr

// input statement: "2 + 3 * 4"
// output statement: Number("2"), Plus("+"), Number("3"), Star("*") and Number("4")

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
	ch := l.input[l.pos]
	for l.pos < len(l.input) {
		// when the function encounters the unwanted expressions, it increments the position by 1
		if (ch == ' ') || (ch == '\t') || (ch == '\r') || (ch == '\n') {
			l.pos++
		}
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
	ch := l.input[l.pos]

	// for the digits before the decimal point
	for (l.pos <= len(l.input) && l.isDigit(ch)) {
		l.pos++
	}

	// for the digits after the decimal point
	if (l.pos <= len(l.input) && ch == '.') {
		l.pos++

		for (l.pos <= len(l.input) && l.isDigit(ch)) {
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
	ch := l.input[l.pos]

	for (l.pos <= len(l.input)) {
		// while the character at that position is a letter, digit or underscore, advance the position
		switch {
			case l.isLetter(ch):
				l.pos++
			case l.isDigit(ch):
				l.pos++
			case ch == '_':
				l.pos++

			default:
				break
		}
	}

	// checking whether any keyword matches or not with the sliced text
	text := l.input[startPos:l.pos]
	var keywords = map[string]TokenKind {
		"let": LET,
		"true": TRUE,
		"false": FALSE,
	}

	// if the sliced text matches with the keywords present in the map, then save it as a keyword
	// else save it as IDENT
	if _, ok := keywords[text]; ok {
		return Token{
			Kind: keywords[text],
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

func (l *lexer) readOperator() Token {
	// if the current position is greater than the length of the whole input statement
	if l.pos >= len(l.input) {
		return Token{
			Kind: ILLEGAL,
			Pos: l.pos,
		}
	}
	
	ch := l.input[l.pos]
	
	operatorKeyword := map[byte]TokenKind{
		'+': PLUS,
		'-': MINUS,
		'*': STAR,
		'/': SLASH,
		'%': PERCENT,
		'^': CARET,
		'!': BANG,
		'=': EQ,
		'<': LT,
		'>': GT,
		// '!=': NEQ,			// review
		// '<': LTE,		// review
		// ">=": GTE,		// review
		// "&&": AND,		// review
		// "||": OR,		// review
	}

	double := map[string]TokenKind{
		"!=": NEQ,
		"<=": LTE,
		">=": GTE,
		"&&": AND,
		"||": OR,
	}

	// if there's an operator, then it will definitely match with a key from the operatorKeyword map
	// if not, then it's defined as ILLEGAL token
	_, ok1 := operatorKeyword[l.input[l.pos]]
	_, ok2 := operatorKeyword[l.input[l.pos+1]]

	// if two consecutive characters are not operators, then it will be mapped easily in the keyword map
	if ok1 && !ok2 {
		return Token{
			Kind: operatorKeyword[ch],
			Text: (string)(ch),
			Pos: l.pos,
		}
	} else if ok1 && ok2 {
		text := l.input[l.pos:l.pos+1]
		return Token{
			Kind: double[text],
			Text: text,
			Pos: l.pos,
		}
	}

	return Token{
		Kind: ILLEGAL,
		Pos: l.pos,
	}
}

// example: "Hello World"
// this function runs when we encounter the opening "
// and closes with the closing "
func (l *lexer) readString() Token {
	startPos := l.pos
	l.pos++

	ch := l.input[l.pos]
	for (l.pos <= len(l.input)) {
		// checking whether the current character is a closing "
		// if it is, then consume the text within ""
		// or if not, then return the token as ILLEGAL
		if ch == '"' {
			// slice the text until the closing '"'
			text := l.input[startPos:l.pos]
			return Token{
				Kind: STRING,
				Text: text,
				Pos: startPos,
			}
		}

		// if the character is not the closing '"', then simply advance the position
		l.pos++
	}

	// return the token as ILLEGAL if the string is not resolved in the for loop
	// meaning there is no closing '"' or the position has gone out of the length of the input
	return Token{
		Kind: ILLEGAL,
		Pos: startPos,
	}
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