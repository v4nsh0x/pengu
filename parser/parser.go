package parser

import (
	"fmt"
	"strconv"

	"github.com/vansh/pengu/ast"
	"github.com/vansh/pengu/lexer"
)

// Parser transforms a token stream into an AST.
type Parser struct {
	tokens  []lexer.Token
	pos     int
	current lexer.Token
}

// New creates a new Parser from a token slice.
func New(tokens []lexer.Token) *Parser {
	p := &Parser{
		tokens: tokens,
		pos:    0,
	}
	if len(tokens) > 0 {
		p.current = tokens[0]
	}
	return p
}

// Parse parses the entire token stream and returns a Program AST node.
func (p *Parser) Parse() (*ast.Program, error) {
	program := &ast.Program{
		Statements: make([]ast.Node, 0),
	}

	for !p.isAtEnd() {
		p.skipNewlines()
		if p.isAtEnd() {
			break
		}
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}

	return program, nil
}

// --- Token helpers ---

func (p *Parser) peek() lexer.Token {
	return p.current
}

func (p *Parser) peekType() lexer.TokenType {
	return p.current.Type
}

func (p *Parser) advance() lexer.Token {
	tok := p.current
	p.pos++
	if p.pos < len(p.tokens) {
		p.current = p.tokens[p.pos]
	} else {
		p.current = lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return tok
}

func (p *Parser) isAtEnd() bool {
	return p.current.Type == lexer.TOKEN_EOF
}

func (p *Parser) expect(tt lexer.TokenType) (lexer.Token, error) {
	if p.current.Type != tt {
		return lexer.Token{}, fmt.Errorf("Syntax Error:\nExpected '%s' but found '%s'\nLine %d, Column %d",
			tt.String(), p.current.Value, p.current.Line, p.current.Column)
	}
	return p.advance(), nil
}

func (p *Parser) match(types ...lexer.TokenType) bool {
	for _, tt := range types {
		if p.current.Type == tt {
			return true
		}
	}
	return false
}

func (p *Parser) skipNewlines() {
	for p.current.Type == lexer.TOKEN_NEWLINE {
		p.advance()
	}
}

func (p *Parser) expectStatementEnd() {
	// Accept newline, EOF, or closing brace as statement terminator
	if p.current.Type == lexer.TOKEN_NEWLINE {
		p.advance()
	}
	// Also skip any additional newlines
	p.skipNewlines()
}

func (p *Parser) lookAhead(offset int) lexer.Token {
	idx := p.pos + offset
	if idx < len(p.tokens) {
		return p.tokens[idx]
	}
	return lexer.Token{Type: lexer.TOKEN_EOF}
}

// --- Statement parsing ---

func (p *Parser) parseStatement() (ast.Node, error) {
	switch p.peekType() {
	case lexer.TOKEN_STORE:
		return p.parseVariableDeclaration()
	case lexer.TOKEN_FN:
		return p.parseFunctionDeclaration()
	case lexer.TOKEN_WHEN:
		return p.parseIfStatement()
	case lexer.TOKEN_REPEAT:
		return p.parseRepeatStatement()
	case lexer.TOKEN_RETURN:
		return p.parseReturnStatement()
	case lexer.TOKEN_SAY:
		return p.parseSayStatement()
	case lexer.TOKEN_USE:
		return p.parseUseStatement()
	case lexer.TOKEN_BREAK:
		return p.parseBreakStatement()
	case lexer.TOKEN_CONTINUE:
		return p.parseContinueStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseVariableDeclaration parses: store name = expr
func (p *Parser) parseVariableDeclaration() (ast.Node, error) {
	tok := p.advance() // consume 'store'
	line := tok.Line

	nameTok, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, fmt.Errorf("Syntax Error:\nExpected variable name after 'store'\nLine %d", line)
	}

	_, err = p.expect(lexer.TOKEN_ASSIGN)
	if err != nil {
		return nil, fmt.Errorf("Syntax Error:\nExpected '=' after variable name '%s'\nLine %d", nameTok.Value, line)
	}

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	p.expectStatementEnd()

	return &ast.VariableDeclaration{
		Name:  nameTok.Value,
		Value: value,
		Line:  line,
	}, nil
}

// parseFunctionDeclaration parses: fn name(params) { body }
func (p *Parser) parseFunctionDeclaration() (ast.Node, error) {
	tok := p.advance() // consume 'fn'
	line := tok.Line

	nameTok, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, fmt.Errorf("Syntax Error:\nExpected function name after 'fn'\nLine %d", line)
	}

	_, err = p.expect(lexer.TOKEN_LPAREN)
	if err != nil {
		return nil, fmt.Errorf("Syntax Error:\nExpected '(' after function name '%s'\nLine %d", nameTok.Value, line)
	}

	params := make([]string, 0)
	for p.peekType() != lexer.TOKEN_RPAREN && !p.isAtEnd() {
		paramTok, err := p.expect(lexer.TOKEN_IDENT)
		if err != nil {
			return nil, fmt.Errorf("Syntax Error:\nExpected parameter name in function '%s'\nLine %d", nameTok.Value, line)
		}
		params = append(params, paramTok.Value)
		if p.peekType() == lexer.TOKEN_COMMA {
			p.advance()
		}
	}

	_, err = p.expect(lexer.TOKEN_RPAREN)
	if err != nil {
		return nil, fmt.Errorf("Syntax Error:\nExpected ')' after parameters in function '%s'\nLine %d", nameTok.Value, line)
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &ast.FunctionDeclaration{
		Name:   nameTok.Value,
		Params: params,
		Body:   body,
		Line:   line,
	}, nil
}

// parseIfStatement parses: when condition { body } otherwise { body }
func (p *Parser) parseIfStatement() (ast.Node, error) {
	tok := p.advance() // consume 'when'
	line := tok.Line

	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	var elseBody []ast.Node
	p.skipNewlines()

	if p.peekType() == lexer.TOKEN_OTHERWISE {
		p.advance() // consume 'otherwise'
		elseBody, err = p.parseBlock()
		if err != nil {
			return nil, err
		}
	}

	return &ast.IfStatement{
		Condition: condition,
		Body:      body,
		ElseBody:  elseBody,
		Line:      line,
	}, nil
}

// parseRepeatStatement parses all repeat forms.
func (p *Parser) parseRepeatStatement() (ast.Node, error) {
	tok := p.advance() // consume 'repeat'
	line := tok.Line

	// Check for: repeat item in collection { body }
	// We look ahead to see if we have: IDENT IN ...
	if p.peekType() == lexer.TOKEN_IDENT && p.lookAhead(1).Type == lexer.TOKEN_IN {
		iteratorTok := p.advance() // consume iterator name
		p.advance()                // consume 'in'

		collection, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		body, err := p.parseBlock()
		if err != nil {
			return nil, err
		}

		return &ast.RepeatStatement{
			Iterator:   iteratorTok.Value,
			Collection: collection,
			Body:       body,
			Line:       line,
		}, nil
	}

	// Parse the expression after 'repeat'
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	// Determine if this is a numeric repeat or a conditional repeat.
	// If the expression is a simple number literal, treat as numeric repeat.
	// Otherwise treat as conditional repeat.
	switch expr.(type) {
	case *ast.NumberLiteral:
		return &ast.RepeatStatement{
			Count: expr,
			Body:  body,
			Line:  line,
		}, nil
	default:
		// Check if it looks like a condition (has comparison/logical ops) or a number expression
		if isConditionExpression(expr) {
			return &ast.RepeatStatement{
				Condition: expr,
				Body:      body,
				Line:      line,
			}, nil
		}
		// Default: treat as count (could be a variable holding a number)
		return &ast.RepeatStatement{
			Count: expr,
			Body:  body,
			Line:  line,
		}, nil
	}
}

// isConditionExpression checks if an expression is likely a condition (comparison/logical).
func isConditionExpression(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.BinaryExpression:
		switch n.Operator {
		case "==", "!=", "<", ">", "<=", ">=", "&&", "||":
			return true
		}
	case *ast.UnaryExpression:
		if n.Operator == "!" {
			return true
		}
	}
	return false
}

func (p *Parser) parseReturnStatement() (ast.Node, error) {
	tok := p.advance() // consume 'return'
	line := tok.Line

	// Check if there's a value to return (not just a newline/EOF/closing brace)
	var value ast.Node
	if !p.isAtEnd() && p.peekType() != lexer.TOKEN_NEWLINE && p.peekType() != lexer.TOKEN_RBRACE {
		var err error
		value, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}

	if value == nil {
		value = &ast.NullLiteral{Line: line}
	}

	p.expectStatementEnd()

	return &ast.ReturnStatement{
		Value: value,
		Line:  line,
	}, nil
}

// parseSayStatement parses: say expr  OR  say(expr)
func (p *Parser) parseSayStatement() (ast.Node, error) {
	tok := p.advance() // consume 'say'
	line := tok.Line

	// If followed by '(', parse as a call expression: say(expr)
	if p.peekType() == lexer.TOKEN_LPAREN {
		p.advance() // consume '('
		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		_, err = p.expect(lexer.TOKEN_RPAREN)
		if err != nil {
			return nil, fmt.Errorf("Syntax Error:\nExpected ')' after say argument\nLine %d", line)
		}
		p.expectStatementEnd()
		return &ast.SayStatement{
			Value: value,
			Line:  line,
		}, nil
	}

	// Otherwise parse as: say expr
	if p.isAtEnd() || p.peekType() == lexer.TOKEN_NEWLINE || p.peekType() == lexer.TOKEN_RBRACE {
		return nil, fmt.Errorf("Syntax Error:\nExpected expression after 'say'\nLine %d", line)
	}

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	p.expectStatementEnd()

	return &ast.SayStatement{
		Value: value,
		Line:  line,
	}, nil
}

func (p *Parser) parseUseStatement() (ast.Node, error) {
	tok := p.advance() // consume 'use'
	line := tok.Line

	modTok, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, fmt.Errorf("Syntax Error:\nExpected module name after 'use'\nLine %d", line)
	}

	p.expectStatementEnd()

	return &ast.UseStatement{
		Module: modTok.Value,
		Line:   line,
	}, nil
}

func (p *Parser) parseBreakStatement() (ast.Node, error) {
	tok := p.advance() // consume 'break'
	p.expectStatementEnd()
	return &ast.BreakStatement{Line: tok.Line}, nil
}

func (p *Parser) parseContinueStatement() (ast.Node, error) {
	tok := p.advance() // consume 'continue'
	p.expectStatementEnd()
	return &ast.ContinueStatement{Line: tok.Line}, nil
}

// parseExpressionStatement handles assignments and bare expression statements.
func (p *Parser) parseExpressionStatement() (ast.Node, error) {
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Check for assignment: ident = expr  or  expr[index] = expr
	if p.peekType() == lexer.TOKEN_ASSIGN {
		p.advance() // consume '='
		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		p.expectStatementEnd()
		return &ast.AssignmentExpression{
			Target: expr,
			Value:  value,
			Line:   p.current.Line,
		}, nil
	}

	p.expectStatementEnd()
	return expr, nil
}

// --- Block parsing ---

func (p *Parser) parseBlock() ([]ast.Node, error) {
	p.skipNewlines()
	_, err := p.expect(lexer.TOKEN_LBRACE)
	if err != nil {
		return nil, fmt.Errorf("Syntax Error:\nExpected '{' to start block\nLine %d, Column %d", p.current.Line, p.current.Column)
	}
	p.skipNewlines()

	stmts := make([]ast.Node, 0)
	for p.peekType() != lexer.TOKEN_RBRACE && !p.isAtEnd() {
		p.skipNewlines()
		if p.peekType() == lexer.TOKEN_RBRACE {
			break
		}
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}

	_, err = p.expect(lexer.TOKEN_RBRACE)
	if err != nil {
		return nil, fmt.Errorf("Syntax Error:\nExpected '}' to close block\nLine %d, Column %d", p.current.Line, p.current.Column)
	}

	return stmts, nil
}

// --- Expression parsing (Pratt parser) ---

// Precedence levels
const (
	PREC_NONE       = iota
	PREC_OR         // ||
	PREC_AND        // &&
	PREC_EQUALITY   // == !=
	PREC_COMPARISON // < > <= >=
	PREC_TERM       // + -
	PREC_FACTOR     // * / %
	PREC_UNARY      // ! -
	PREC_CALL       // () [] .
)

func (p *Parser) parseExpression() (ast.Node, error) {
	return p.parseOr()
}

func (p *Parser) parseOr() (ast.Node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.peekType() == lexer.TOKEN_OR {
		op := p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpression{
			Left:     left,
			Operator: op.Value,
			Right:    right,
			Line:     op.Line,
		}
	}
	return left, nil
}

func (p *Parser) parseAnd() (ast.Node, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}

	for p.peekType() == lexer.TOKEN_AND {
		op := p.advance()
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpression{
			Left:     left,
			Operator: op.Value,
			Right:    right,
			Line:     op.Line,
		}
	}
	return left, nil
}

func (p *Parser) parseEquality() (ast.Node, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.TOKEN_EQ, lexer.TOKEN_NEQ) {
		op := p.advance()
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpression{
			Left:     left,
			Operator: op.Value,
			Right:    right,
			Line:     op.Line,
		}
	}
	return left, nil
}

func (p *Parser) parseComparison() (ast.Node, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.TOKEN_LT, lexer.TOKEN_GT, lexer.TOKEN_LTE, lexer.TOKEN_GTE) {
		op := p.advance()
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpression{
			Left:     left,
			Operator: op.Value,
			Right:    right,
			Line:     op.Line,
		}
	}
	return left, nil
}

func (p *Parser) parseTerm() (ast.Node, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.TOKEN_PLUS, lexer.TOKEN_MINUS) {
		op := p.advance()
		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpression{
			Left:     left,
			Operator: op.Value,
			Right:    right,
			Line:     op.Line,
		}
	}
	return left, nil
}

func (p *Parser) parseFactor() (ast.Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.TOKEN_STAR, lexer.TOKEN_SLASH, lexer.TOKEN_PERCENT) {
		op := p.advance()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpression{
			Left:     left,
			Operator: op.Value,
			Right:    right,
			Line:     op.Line,
		}
	}
	return left, nil
}

func (p *Parser) parseUnary() (ast.Node, error) {
	if p.match(lexer.TOKEN_NOT, lexer.TOKEN_MINUS) {
		op := p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpression{
			Operator: op.Value,
			Operand:  operand,
			Line:     op.Line,
		}, nil
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() (ast.Node, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for {
		switch p.peekType() {
		case lexer.TOKEN_LPAREN:
			// Function call
			p.advance() // consume '('
			args := make([]ast.Node, 0)
			for p.peekType() != lexer.TOKEN_RPAREN && !p.isAtEnd() {
				arg, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				if p.peekType() == lexer.TOKEN_COMMA {
					p.advance()
				}
			}
			_, err := p.expect(lexer.TOKEN_RPAREN)
			if err != nil {
				return nil, err
			}
			expr = &ast.CallExpression{
				Callee:    expr,
				Arguments: args,
				Line:      p.current.Line,
			}

		case lexer.TOKEN_LBRACKET:
			// Index access
			p.advance() // consume '['
			index, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			_, err = p.expect(lexer.TOKEN_RBRACKET)
			if err != nil {
				return nil, err
			}
			expr = &ast.IndexExpression{
				Object: expr,
				Index:  index,
				Line:   p.current.Line,
			}

		case lexer.TOKEN_DOT:
			// Member access
			p.advance() // consume '.'
			propTok, err := p.expect(lexer.TOKEN_IDENT)
			if err != nil {
				return nil, fmt.Errorf("Syntax Error:\nExpected property name after '.'\nLine %d", p.current.Line)
			}
			expr = &ast.MemberExpression{
				Object:   expr,
				Property: propTok.Value,
				Line:     propTok.Line,
			}

		default:
			return expr, nil
		}
	}
}

func (p *Parser) parsePrimary() (ast.Node, error) {
	tok := p.peek()

	switch tok.Type {
	case lexer.TOKEN_INT:
		p.advance()
		val, err := strconv.ParseInt(tok.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Syntax Error:\nInvalid integer '%s'\nLine %d", tok.Value, tok.Line)
		}
		return &ast.NumberLiteral{Value: float64(val), IsInt: true, Line: tok.Line}, nil

	case lexer.TOKEN_FLOAT:
		p.advance()
		val, err := strconv.ParseFloat(tok.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("Syntax Error:\nInvalid number '%s'\nLine %d", tok.Value, tok.Line)
		}
		return &ast.NumberLiteral{Value: val, IsInt: false, Line: tok.Line}, nil

	case lexer.TOKEN_STRING:
		p.advance()
		return &ast.StringLiteral{Value: tok.Value, Line: tok.Line}, nil

	case lexer.TOKEN_TRUE:
		p.advance()
		return &ast.BooleanLiteral{Value: true, Line: tok.Line}, nil

	case lexer.TOKEN_FALSE:
		p.advance()
		return &ast.BooleanLiteral{Value: false, Line: tok.Line}, nil

	case lexer.TOKEN_NULL:
		p.advance()
		return &ast.NullLiteral{Line: tok.Line}, nil

	case lexer.TOKEN_IDENT:
		p.advance()
		return &ast.Identifier{Name: tok.Value, Line: tok.Line}, nil

	case lexer.TOKEN_LPAREN:
		// Grouped expression
		p.advance() // consume '('
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		_, err = p.expect(lexer.TOKEN_RPAREN)
		if err != nil {
			return nil, fmt.Errorf("Syntax Error:\nExpected ')' to close grouped expression\nLine %d", tok.Line)
		}
		return expr, nil

	case lexer.TOKEN_LBRACKET:
		// Array literal
		return p.parseArrayLiteral()

	case lexer.TOKEN_LBRACE:
		// Object literal
		return p.parseObjectLiteral()

	case lexer.TOKEN_FN:
		// Anonymous function (lambda): fn(params) { body }
		return p.parseAnonymousFunction()

	default:
		return nil, fmt.Errorf("Syntax Error:\nUnexpected token '%s'\nLine %d, Column %d",
			tok.Value, tok.Line, tok.Column)
	}
}

func (p *Parser) parseArrayLiteral() (ast.Node, error) {
	tok := p.advance() // consume '['
	line := tok.Line

	elements := make([]ast.Node, 0)
	p.skipNewlines()
	for p.peekType() != lexer.TOKEN_RBRACKET && !p.isAtEnd() {
		p.skipNewlines()
		elem, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)
		p.skipNewlines()
		if p.peekType() == lexer.TOKEN_COMMA {
			p.advance()
		}
		p.skipNewlines()
	}

	_, err := p.expect(lexer.TOKEN_RBRACKET)
	if err != nil {
		return nil, fmt.Errorf("Syntax Error:\nExpected ']' to close array\nLine %d", line)
	}

	return &ast.ArrayLiteral{
		Elements: elements,
		Line:     line,
	}, nil
}

func (p *Parser) parseObjectLiteral() (ast.Node, error) {
	tok := p.advance() // consume '{'
	line := tok.Line

	keys := make([]ast.Node, 0)
	values := make([]ast.Node, 0)

	p.skipNewlines()
	for p.peekType() != lexer.TOKEN_RBRACE && !p.isAtEnd() {
		p.skipNewlines()
		// Key can be a string or identifier
		var key ast.Node
		if p.peekType() == lexer.TOKEN_STRING {
			keyTok := p.advance()
			key = &ast.StringLiteral{Value: keyTok.Value, Line: keyTok.Line}
		} else if p.peekType() == lexer.TOKEN_IDENT {
			keyTok := p.advance()
			key = &ast.StringLiteral{Value: keyTok.Value, Line: keyTok.Line}
		} else {
			return nil, fmt.Errorf("Syntax Error:\nExpected string or identifier as object key\nLine %d", p.current.Line)
		}
		keys = append(keys, key)

		_, err := p.expect(lexer.TOKEN_COLON)
		if err != nil {
			return nil, fmt.Errorf("Syntax Error:\nExpected ':' after object key\nLine %d", p.current.Line)
		}

		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		values = append(values, value)

		p.skipNewlines()
		if p.peekType() == lexer.TOKEN_COMMA {
			p.advance()
		}
		p.skipNewlines()
	}

	_, err := p.expect(lexer.TOKEN_RBRACE)
	if err != nil {
		return nil, fmt.Errorf("Syntax Error:\nExpected '}' to close object\nLine %d", line)
	}

	return &ast.ObjectLiteral{
		Keys:   keys,
		Values: values,
		Line:   line,
	}, nil
}

func (p *Parser) parseAnonymousFunction() (ast.Node, error) {
	tok := p.advance() // consume 'fn'
	line := tok.Line

	_, err := p.expect(lexer.TOKEN_LPAREN)
	if err != nil {
		return nil, fmt.Errorf("Syntax Error:\nExpected '(' for anonymous function\nLine %d", line)
	}

	params := make([]string, 0)
	for p.peekType() != lexer.TOKEN_RPAREN && !p.isAtEnd() {
		paramTok, err := p.expect(lexer.TOKEN_IDENT)
		if err != nil {
			return nil, fmt.Errorf("Syntax Error:\nExpected parameter name\nLine %d", line)
		}
		params = append(params, paramTok.Value)
		if p.peekType() == lexer.TOKEN_COMMA {
			p.advance()
		}
	}

	_, err = p.expect(lexer.TOKEN_RPAREN)
	if err != nil {
		return nil, err
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &ast.FunctionDeclaration{
		Name:   "<anonymous>",
		Params: params,
		Body:   body,
		Line:   line,
	}, nil
}
