package sexpr

import (
	"errors"
)

// ErrParser is the error value returned by the Parser if the string is not a
// valid term.
// See also https://golang.org/pkg/errors/#New
// and // https://golang.org/pkg/builtin/#error
var ErrParser = errors.New("parser error")

//
// <sexpr>       ::= <atom> | <pars> | QUOTE <sexpr>
// <atom>        ::= NUMBER | SYMBOL
// <pars>        ::= LPAR <dotted_list> RPAR | LPAR <proper_list> RPAR
// <dotted_list> ::= <proper_list> <sexpr> DOT <sexpr>
// <proper_list> ::= <sexpr> <proper_list> | \epsilon
//

//
// <sexpr>       ::= NUMBER | SYMBOL | LPAR <pars> RPAR | QUOTE <sexpr>
// <pars>        ::= <sexpr> <pars'>
// <pars'>		 ::= DOT <sexpr> | <pars> | \epsilon
//

// 					|			FIRST							 |			FOLLOW									||
// 					|											 |													||
// <sexpr>			|	NUMBER, SYMBOL, LPAR, QUOTE				 |	NUMBER, SYMBOL, LPAR, QUOTE, $					||
// <pars>			|	NUMBER, SYMBOL, LPAR, QUOTE				 |	RPAR											||
// <pars'>			|	DOT, NUMBER, SYMBOL, LPAR, QUOTE,\epsilon|	RPAR											||

// 					|		NUMBER									|		SYMBOL									|			QUOTE							|	$										||LPAR											| 	RPAR 					|	DOT
// 					|												|												|											|											||												|							|
// <sexpr>			| <sexpr> -> NUMBER								|	<sexpr> -> SYMBOL							| <sexpr> -> QUOTE <sexpr>					|											||	<sexpr> -> LPAR<pars'>						|							|
// 					|												|												|											|											||												|							|
// <pars>			| <pars> -> <sexpr> <pars'>		  				|	<pars> -> <sexpr> <pars'>					| <pars> -> <sexpr> <pars'>					| 											||	<pars> -> <sexpr> <pars'>					| <pars> -> \epsilon		|
// 					|												|												|											|											||												|							|
// <pars'>			| <pars'> -> DOT <sexpr> | <list'>				| 	<pars'> -> DOT <sexpr> | <list'>			| <pars'> -> DOT <sexpr> | <list'>			|											|| <pars'> -> DOT <sexpr> | <list'>				| <pars'> -> \epsilon		| <pars'> -> DOT <sexpr>| <list'>
// 					|												|												|											|											||												|							|

var LPAR_EXST = false

type Parser interface {
	Parse(string) (*SExpr, error)
}

type parserStruct struct {
	lex     *lexer
	peekTok *token
}

func (p *parserStruct) nextToken() (*token, error) {
	if tok := p.peekTok; tok != nil {
		p.peekTok = nil
		return tok, nil
	}

	tok, err := p.lex.next()
	if err != nil {
		return nil, ErrParser
	}

	return tok, nil
}

// Helper function which puts a token back as the next token.
func (p *parserStruct) backToken(tok *token) {
	p.peekTok = tok
}

// Helper function to peek the next token.
func (p *parserStruct) peekToken() (*token, error) {
	tok, err := p.nextToken()
	if err != nil {
		return nil, ErrParser
	}

	p.backToken(tok)

	return tok, nil
}

func NewParser() Parser {
	return &parserStruct{}
}

func (p *parserStruct) Parse(input string) (*SExpr, error) {
	LPAR_EXST = false
	p.lex = newLexer(input)
	p.peekTok = nil

	tok, err := p.nextToken()
	if err != nil {
		return nil, ErrParser
	}

	if len(input) > 1 {
		if tok.typ != tokenLpar && tok.typ != tokenQuote && tok.typ != tokenNumber {
			return nil, ErrParser
		}
	}

	tok1, err := p.nextToken()
	if err != nil {
		return nil, ErrParser
	}

	if tok.typ == tokenLpar && tok1.typ == tokenRpar {
		tok1, err := p.peekToken()
		if err != nil {
			return nil, ErrParser
		}
		if tok1.typ != tokenEOF {
			return nil, ErrParser
		}
	}

	p.lex = newLexer(input)
	p.peekTok = nil

	expr, err := p.SExpr()
	if err != nil {
		return nil, ErrParser
	}

	tok, err = p.peekToken()

	if tok.typ != tokenEOF {
		return nil, ErrParser
	}

	return expr, nil
}

//parses  <sexpr>  ::= NUMBER | SYMBOL | LPAR <pars> RPAR | QUOTE <sexpr>
// <sexpr>	|	First - NUMBER, SYMBOL, LPAR, QUOTE	|	Follow- NUMBER, SYMBOL, LPAR, QUOTE, $	||
func (p *parserStruct) SExpr() (*SExpr, error) {
	tok, err := p.nextToken()

	tok1, err := p.peekToken()
	if err != nil {
		return nil, ErrParser
	}

	//check for firsts of sexpr
	switch tok.typ {
	case tokenNumber: //at token TokenNumber
		if tok1.typ == tokenRpar && LPAR_EXST == false {
			return nil, ErrParser
		}
		return mkNumber(tok.num), nil

	case tokenSymbol: //at token TokenSymbol
		return mkSymbol(tok.literal), nil

	case tokenLpar: //at token LPAR
		LPAR_EXST = true
		expr, err := p.Pars()
		if err != nil {
			return nil, ErrParser
		}

		tok, err := p.nextToken() //At token RPAR

		_, err = p.peekToken() //Check invalid characters after RPAR
		if err != nil {
			return nil, ErrParser
		}

		if tok.typ == tokenRpar {
			return expr, nil
		} else {
			return nil, ErrParser
		}

	case tokenQuote: //at token TokenQuote
		if tok1.typ == tokenRpar {
			return nil, ErrParser
		}

		tok2 := tok
		expr, err := p.SExpr()
		if err != nil {
			return nil, ErrParser
		}
		list := mkConsCell(expr, mkNil())
		return mkConsCell(mkAtom(tok2), list), nil

	default:
		return nil, ErrParser
	}
}

// <pars>			|	NUMBER, SYMBOL, LPAR, QUOTE				|	RPAR											||
// <pars>        ::= <sexpr> <pars'>
func (p *parserStruct) Pars() (*SExpr, error) {
	tok, err := p.peekToken() //At token DOT/epsilon/sexpr

	if tok.typ == tokenRpar {
		return mkNil(), nil
	}

	expr1, err := p.SExpr()
	if err != nil {
		return nil, ErrParser
	}

	expr2, err := p.Pars2()
	if err != nil {
		return nil, ErrParser
	}

	return mkConsCell(expr1, expr2), nil
}

// <pars'>		 ::= DOT <sexpr> | <pars> | \epsilon
// <pars'>			|	first - DOT, NUMBER, SYMBOL, LPAR, QUOTE,\epsilon| follow-	RPAR
func (p *parserStruct) Pars2() (*SExpr, error) {
	tok, _ := p.peekToken() //At token DOT/epsilon/sexpr

	switch tok.typ {
	case tokenDot:
		_, err := p.nextToken() //At token DOT
		expr, err := p.SExpr()
		if err != nil {
			return nil, ErrParser
		}
		return expr, nil

	case tokenRpar:
		return mkNil(), nil

	case tokenNumber:
		expr, _ := p.Pars()
		return expr, nil

	case tokenSymbol:
		expr, _ := p.Pars()
		return expr, nil

	case tokenQuote:
		expr, _ := p.Pars()
		return expr, nil

	case tokenLpar:
		expr, _ := p.Pars()
		return expr, nil

	default:
		return nil, ErrParser
	}
}
