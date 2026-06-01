package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

// Lexer tokenizes Pengu source code.
type Lexer struct {
	source  []rune
	pos     int
	line    int
	col     int
	tokens  []Token
}

// New creates a new Lexer for the given source code.
func New(source string) *Lexer {
	return &Lexer{
		source: []rune(source),
		pos:    0,
		line:   1,
		col:    1,
	}
}

// Tokenize scans the entire source and returns all tokens.
func (l *Lexer) Tokenize() ([]Token, error) {
	for {
		tok, err := l.nextToken()
		if err != nil {
			return nil, err
		}
		l.tokens = append(l.tokens, tok)
		if tok.Type == TOKEN_EOF {
			break
		}
	}
	return l.tokens, nil
}

// peek returns the current character without advancing.
func (l *Lexer) peek() rune {
	if l.pos >= len(l.source) {
		return 0
	}
	return l.source[l.pos]
}

// peekNext returns the next character without advancing.
func (l *Lexer) peekNext() rune {
	if l.pos+1 >= len(l.source) {
		return 0
	}
	return l.source[l.pos+1]
}

// advance moves forward one character and returns it.
func (l *Lexer) advance() rune {
	ch := l.source[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

// makeToken creates a token at the current position.
func (l *Lexer) makeToken(tt TokenType, value string, line, col int) Token {
	return Token{
		Type:   tt,
		Value:  value,
		Line:   line,
		Column: col,
	}
}

// skipWhitespace skips spaces and tabs (NOT newlines).
func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.source) {
		ch := l.peek()
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.advance()
		} else {
			break
		}
	}
}

// skipLineComment skips a // comment.
func (l *Lexer) skipLineComment() {
	for l.pos < len(l.source) && l.peek() != '\n' {
		l.advance()
	}
}

// skipBlockComment skips a /* ... */ comment.
func (l *Lexer) skipBlockComment() error {
	startLine := l.line
	l.advance() // skip /
	l.advance() // skip *
	for l.pos < len(l.source) {
		if l.peek() == '*' && l.peekNext() == '/' {
			l.advance() // skip *
			l.advance() // skip /
			return nil
		}
		l.advance()
	}
	return fmt.Errorf("Syntax Error:\nUnterminated block comment\nStarted at line %d", startLine)
}

// nextToken scans and returns the next token.
func (l *Lexer) nextToken() (Token, error) {
	l.skipWhitespace()

	if l.pos >= len(l.source) {
		return l.makeToken(TOKEN_EOF, "", l.line, l.col), nil
	}

	ch := l.peek()
	line := l.line
	col := l.col

	// Newlines
	if ch == '\n' {
		l.advance()
		// Collapse multiple newlines
		for l.pos < len(l.source) && (l.peek() == '\n' || l.peek() == '\r' || l.peek() == ' ' || l.peek() == '\t') {
			l.advance()
		}
		return l.makeToken(TOKEN_NEWLINE, "\n", line, col), nil
	}

	// Comments
	if ch == '/' {
		if l.peekNext() == '/' {
			l.skipLineComment()
			return l.nextToken()
		}
		if l.peekNext() == '*' {
			if err := l.skipBlockComment(); err != nil {
				return Token{}, err
			}
			return l.nextToken()
		}
	}

	// Strings
	if ch == '"' {
		return l.readString(false)
	}

	// F-Strings and Identifiers
	if ch == 'f' && l.peekNext() == '"' {
		l.advance() // consume 'f'
		return l.readString(true)
	}

	// Numbers
	if unicode.IsDigit(ch) {
		return l.readNumber()
	}

	// Identifiers and keywords
	if unicode.IsLetter(ch) || ch == '_' {
		return l.readIdentifier()
	}

	// Two-character operators
	switch {
	case ch == '=' && l.peekNext() == '=':
		l.advance()
		l.advance()
		return l.makeToken(TOKEN_EQ, "==", line, col), nil
	case ch == '!' && l.peekNext() == '=':
		l.advance()
		l.advance()
		return l.makeToken(TOKEN_NEQ, "!=", line, col), nil
	case ch == '<' && l.peekNext() == '=':
		l.advance()
		l.advance()
		return l.makeToken(TOKEN_LTE, "<=", line, col), nil
	case ch == '>' && l.peekNext() == '=':
		l.advance()
		l.advance()
		return l.makeToken(TOKEN_GTE, ">=", line, col), nil
	case ch == '&' && l.peekNext() == '&':
		l.advance()
		l.advance()
		return l.makeToken(TOKEN_AND, "&&", line, col), nil
	case ch == '|' && l.peekNext() == '|':
		l.advance()
		l.advance()
		return l.makeToken(TOKEN_OR, "||", line, col), nil
	}

	// Single-character tokens
	l.advance()
	switch ch {
	case '+':
		return l.makeToken(TOKEN_PLUS, "+", line, col), nil
	case '-':
		return l.makeToken(TOKEN_MINUS, "-", line, col), nil
	case '*':
		return l.makeToken(TOKEN_STAR, "*", line, col), nil
	case '/':
		return l.makeToken(TOKEN_SLASH, "/", line, col), nil
	case '%':
		return l.makeToken(TOKEN_PERCENT, "%", line, col), nil
	case '=':
		return l.makeToken(TOKEN_ASSIGN, "=", line, col), nil
	case '<':
		return l.makeToken(TOKEN_LT, "<", line, col), nil
	case '>':
		return l.makeToken(TOKEN_GT, ">", line, col), nil
	case '!':
		return l.makeToken(TOKEN_NOT, "!", line, col), nil
	case '(':
		return l.makeToken(TOKEN_LPAREN, "(", line, col), nil
	case ')':
		return l.makeToken(TOKEN_RPAREN, ")", line, col), nil
	case '{':
		return l.makeToken(TOKEN_LBRACE, "{", line, col), nil
	case '}':
		return l.makeToken(TOKEN_RBRACE, "}", line, col), nil
	case '[':
		return l.makeToken(TOKEN_LBRACKET, "[", line, col), nil
	case ']':
		return l.makeToken(TOKEN_RBRACKET, "]", line, col), nil
	case ',':
		return l.makeToken(TOKEN_COMMA, ",", line, col), nil
	case ':':
		return l.makeToken(TOKEN_COLON, ":", line, col), nil
	case '.':
		return l.makeToken(TOKEN_DOT, ".", line, col), nil
	}

	return Token{}, fmt.Errorf("Syntax Error:\nUnexpected character '%c'\nLine %d, Column %d", ch, line, col)
}

// readString reads a string literal enclosed in double quotes.
func (l *Lexer) readString(isFString bool) (Token, error) {
	line := l.line
	col := l.col
	l.advance() // skip opening "

	var sb strings.Builder
	for l.pos < len(l.source) {
		ch := l.peek()
		if ch == '"' {
			l.advance() // skip closing "
			tokType := TOKEN_STRING
			if isFString {
				tokType = TOKEN_FSTRING
			}
			return l.makeToken(tokType, sb.String(), line, col), nil
		}
		if ch == '\\' {
			l.advance()
			if l.pos >= len(l.source) {
				return Token{}, fmt.Errorf("Syntax Error:\nUnterminated string escape\nLine %d", line)
			}
			escaped := l.advance()
			switch escaped {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case '\\':
				sb.WriteRune('\\')
			case '"':
				sb.WriteRune('"')
			case 'r':
				sb.WriteRune('\r')
			case 'e':
				sb.WriteRune('\033') // ASCII escape character for ANSI colors
			default:
				sb.WriteRune('\\')
				sb.WriteRune(escaped)
			}
			continue
		}
		if ch == '\n' {
			return Token{}, fmt.Errorf("Syntax Error:\nUnterminated string literal\nLine %d", line)
		}
		sb.WriteRune(l.advance())
	}

	return Token{}, fmt.Errorf("Syntax Error:\nUnterminated string literal\nLine %d", line)
}

// readNumber reads an integer or floating-point number.
func (l *Lexer) readNumber() (Token, error) {
	line := l.line
	col := l.col
	var sb strings.Builder
	isFloat := false

	for l.pos < len(l.source) {
		ch := l.peek()
		if unicode.IsDigit(ch) {
			sb.WriteRune(l.advance())
		} else if ch == '.' && !isFloat {
			// Check that next char is a digit (not a method call)
			if l.peekNext() != 0 && unicode.IsDigit(l.peekNext()) {
				isFloat = true
				sb.WriteRune(l.advance()) // the dot
			} else {
				break
			}
		} else {
			break
		}
	}

	if isFloat {
		return l.makeToken(TOKEN_FLOAT, sb.String(), line, col), nil
	}
	return l.makeToken(TOKEN_INT, sb.String(), line, col), nil
}

// readIdentifier reads an identifier or keyword.
func (l *Lexer) readIdentifier() (Token, error) {
	line := l.line
	col := l.col
	var sb strings.Builder

	for l.pos < len(l.source) {
		ch := l.peek()
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
			sb.WriteRune(l.advance())
		} else {
			break
		}
	}

	word := sb.String()
	tt := LookupIdent(word)
	return l.makeToken(tt, word, line, col), nil
}
