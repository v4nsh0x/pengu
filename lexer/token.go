package lexer

// TokenType represents the type of a lexical token.
type TokenType int

const (
	// Special tokens
	TOKEN_EOF TokenType = iota
	TOKEN_ILLEGAL

	// Literals
	TOKEN_IDENT   // variable names, function names
	TOKEN_INT     // integer literals
	TOKEN_FLOAT   // float literals
	TOKEN_STRING  // string literals
	TOKEN_FSTRING // f-string literals
	TOKEN_TRUE    // true
	TOKEN_FALSE   // false
	TOKEN_NULL    // null

	// Keywords
	TOKEN_STORE     // store
	TOKEN_SAY       // say
	TOKEN_FN        // fn
	TOKEN_RETURN    // return
	TOKEN_WHEN      // when
	TOKEN_OTHERWISE // otherwise
	TOKEN_REPEAT    // repeat
	TOKEN_IN        // in
	TOKEN_USE       // use
	TOKEN_BREAK     // break
	TOKEN_CONTINUE  // continue
	TOKEN_TRY       // try
	TOKEN_CATCH     // catch
	TOKEN_SPAWN     // spawn
	TOKEN_AWAIT     // await

	// Operators
	TOKEN_PLUS     // +
	TOKEN_MINUS    // -
	TOKEN_STAR     // *
	TOKEN_SLASH    // /
	TOKEN_PERCENT  // %
	TOKEN_ASSIGN   // =
	TOKEN_EQ       // ==
	TOKEN_NEQ      // !=
	TOKEN_LT       // <
	TOKEN_GT       // >
	TOKEN_LTE      // <=
	TOKEN_GTE      // >=
	TOKEN_AND      // &&
	TOKEN_OR       // ||
	TOKEN_NOT      // !

	// Delimiters
	TOKEN_LPAREN   // (
	TOKEN_RPAREN   // )
	TOKEN_LBRACE   // {
	TOKEN_RBRACE   // }
	TOKEN_LBRACKET // [
	TOKEN_RBRACKET // ]
	TOKEN_COMMA    // ,
	TOKEN_COLON    // :
	TOKEN_DOT      // .
	TOKEN_NEWLINE  // newline (statement separator)
)

// Token represents a lexical token with its type, value, and position.
type Token struct {
	Type    TokenType
	Value   string
	Line    int
	Column  int
}

// keywords maps keyword strings to their token types.
var keywords = map[string]TokenType{
	"store":     TOKEN_STORE,
	"say":       TOKEN_SAY,
	"fn":        TOKEN_FN,
	"return":    TOKEN_RETURN,
	"when":      TOKEN_WHEN,
	"otherwise": TOKEN_OTHERWISE,
	"repeat":    TOKEN_REPEAT,
	"in":        TOKEN_IN,
	"use":       TOKEN_USE,
	"true":      TOKEN_TRUE,
	"false":     TOKEN_FALSE,
	"null":      TOKEN_NULL,
	"break":     TOKEN_BREAK,
	"continue":  TOKEN_CONTINUE,
	"try":       TOKEN_TRY,
	"catch":     TOKEN_CATCH,
	"spawn":     TOKEN_SPAWN,
	"await":     TOKEN_AWAIT,
}

// LookupIdent checks if an identifier is a keyword and returns the appropriate token type.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TOKEN_IDENT
}

// String returns a human-readable name for the token type.
func (t TokenType) String() string {
	switch t {
	case TOKEN_EOF:
		return "EOF"
	case TOKEN_ILLEGAL:
		return "ILLEGAL"
	case TOKEN_IDENT:
		return "IDENT"
	case TOKEN_INT:
		return "INT"
	case TOKEN_FLOAT:
		return "FLOAT"
	case TOKEN_STRING:
		return "STRING"
	case TOKEN_FSTRING:
		return "FSTRING"
	case TOKEN_TRUE:
		return "TRUE"
	case TOKEN_FALSE:
		return "FALSE"
	case TOKEN_NULL:
		return "NULL"
	case TOKEN_STORE:
		return "STORE"
	case TOKEN_SAY:
		return "SAY"
	case TOKEN_FN:
		return "FN"
	case TOKEN_RETURN:
		return "RETURN"
	case TOKEN_WHEN:
		return "WHEN"
	case TOKEN_OTHERWISE:
		return "OTHERWISE"
	case TOKEN_REPEAT:
		return "REPEAT"
	case TOKEN_IN:
		return "IN"
	case TOKEN_USE:
		return "USE"
	case TOKEN_BREAK:
		return "BREAK"
	case TOKEN_CONTINUE:
		return "CONTINUE"
	case TOKEN_TRY:
		return "TRY"
	case TOKEN_CATCH:
		return "CATCH"
	case TOKEN_SPAWN:
		return "SPAWN"
	case TOKEN_AWAIT:
		return "AWAIT"
	case TOKEN_PLUS:
		return "+"
	case TOKEN_MINUS:
		return "-"
	case TOKEN_STAR:
		return "*"
	case TOKEN_SLASH:
		return "/"
	case TOKEN_PERCENT:
		return "%"
	case TOKEN_ASSIGN:
		return "="
	case TOKEN_EQ:
		return "=="
	case TOKEN_NEQ:
		return "!="
	case TOKEN_LT:
		return "<"
	case TOKEN_GT:
		return ">"
	case TOKEN_LTE:
		return "<="
	case TOKEN_GTE:
		return ">="
	case TOKEN_AND:
		return "&&"
	case TOKEN_OR:
		return "||"
	case TOKEN_NOT:
		return "!"
	case TOKEN_LPAREN:
		return "("
	case TOKEN_RPAREN:
		return ")"
	case TOKEN_LBRACE:
		return "{"
	case TOKEN_RBRACE:
		return "}"
	case TOKEN_LBRACKET:
		return "["
	case TOKEN_RBRACKET:
		return "]"
	case TOKEN_COMMA:
		return ","
	case TOKEN_COLON:
		return ":"
	case TOKEN_DOT:
		return "."
	case TOKEN_NEWLINE:
		return "NEWLINE"
	default:
		return "UNKNOWN"
	}
}
