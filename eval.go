package sexpr

import (
	"errors"
	"math/big"
	//"math/big" // You will need to use this package in your implementation.
)

// ErrEval is the error value returned by the Evaluator if the contains
// an invalid token.
// See also https://golang.org/pkg/errors/#New
// and // https://golang.org/pkg/builtin/#error
var ErrEval = errors.New("eval error")

func (expr *SExpr) Eval() (*SExpr, error) {
	if expr.car == nil && expr.cdr == nil {
		if (expr.atom.literal == "+" || expr.atom.literal == "*") && expr.car == nil && expr.cdr == nil {
			return nil, ErrEval
		} else if expr.isNumber() {
			return mkNumber(expr.atom.num), nil
		}
		return nil, ErrEval
	}

	if expr.car.atom.literal == "QUOTE" {
		if expr.cdr.car == nil {
			return mkNil(), ErrEval
		}
		return expr.quote()

	} else if expr.car.atom.literal == "CAR" {
		if expr.cdr.isNil() {
			return mkNil(), ErrEval
		}
		return expr.carr()
	} else if expr.car.atom.literal == "CDR" {
		if expr.cdr.isNil() {
			return mkNil(), ErrEval
		}
		return expr.cdrr()
	} else if expr.car.atom.literal == "CONS" {

		if expr.cdr.isNil() {

			return mkNil(), ErrEval

		} else {

			result, _ := expr.cons()

			if result.isNil() {
				return mkNil(), ErrEval
			}
			if result.car.car == nil {
				return expr.cons()
			}
			if result.car.car.isSexpr() {
				ex, _ := result.car.Eval()
				ex1 := mkConsCell(ex, result.cdr)
				return ex1, nil

			}

		}

	} else if expr.car.atom.literal == "+" {
		if expr.cdr.isNil() {

			return mkNumber(big.NewInt(0)), nil

		}
		if expr.cdr.car.isSymbol() {
			return mkNil(), ErrEval
		}

		if expr.cdr.car.isConsCell() {
			if expr.cdr.car.car.atom.literal == "QUOTE" {
				return mkNil(), ErrEval
			}
		}

		result, _ := expr.add()
		return mkNumber(result.atom.num), nil

	} else if expr.car.atom.literal == "*" {
		if expr.cdr.isNil() {
			return mkNumber(big.NewInt(1)), nil
		}

		if expr.cdr.car.isSymbol() {
			return mkNil(), ErrEval
		}

		if expr.cdr.car.isConsCell() {
			if expr.cdr.car.car.atom.literal == "QUOTE" {
				return mkNil(), ErrEval
			}
		}
		result, _ := expr.mult()
		return mkNumber(result.atom.num), nil

	} else if expr.car.atom.literal == "LENGTH" {

		if expr.cdr.isNil() {
			return mkNil(), ErrEval
		}
		if expr.cdr.car.isNumber() {
			return mkNil(), ErrEval
		}
		if expr.cdr.car.isSymbol() {
			return mkNil(), ErrEval
		}
		if expr.cdr.car.car.atom.literal == "QUOTE" {
			expr2, _ := expr.cdr.car.Eval()

			if expr2.isNil() {
				return mkNumber(big.NewInt(0)), nil
			}

			if expr2.isAtom() {
				return mkNil(), ErrEval
			}

			if expr2.cdr.isAtom() && !expr2.cdr.isNil() {
				return mkNil(), ErrEval
			}

			var len int64 = 0
			result := expr2.length(len)
			return mkNumber(big.NewInt(result)), nil

		} else if expr.cdr.car.car.atom.literal == "CDR" {
			evaluate, _ := expr.cdr.car.Eval()

			var len int64 = 0
			result := evaluate.car.length(len)

			return mkNumber(big.NewInt(result)), nil

		} else if expr.cdr.car.car.atom.literal == "CAR" {
			evaluate, _ := expr.cdr.car.Eval()

			var len int64 = 0
			result := evaluate.length(len)
			return mkNumber(big.NewInt(result)), nil

		} else if expr.cdr.car.car.atom.literal == "CONS" {
			evaluate, _ := expr.cdr.car.Eval()

			var len int64 = 0
			result := evaluate.length(len)
			return mkNumber(big.NewInt(result)), nil

		}

	} else if expr.car.atom.literal == "ATOM" {
		if expr.cdr.car == nil {
			return nil, ErrEval
		}

		if !expr.cdr.cdr.isNil() && expr.cdr.car.isAtom() && expr.cdr.car.atom.literal != "NIL" {
			return mkNil(), ErrEval
		}

		if expr.cdr.car.isConsCell() {
			if expr.cdr.car.car == nil {
				return mkSymbolTrue(), nil
			}

			if expr.cdr.car.car.isSexpr() {
				temp, _ := expr.cdr.car.Eval()
				if temp.isAtom() {
					return mkSymbolTrue(), nil
				}
				return mkNil(), nil
			}

		} else if expr.cdr.car.atom.literal == "NIL" {
			return mkSymbolTrue(), nil
		} else if expr.cdr.car.isNumber() {
			return mkSymbolTrue(), nil
		}

		return nil, ErrEval
	} else if expr.car.atom.literal == "LISTP" {
		return expr.list()
	} else if expr.car.atom.literal == "ZEROP" {
		if expr.cdr.car == nil || expr.cdr.car.isNil() {
			return nil, ErrEval
		}
		if expr.cdr.car.isSymbol() && expr.cdr.cdr.isNil() {

			return mkNil(), ErrEval
		}
		if !expr.cdr.cdr.isNil() && expr.cdr.car.isAtom() && expr.cdr.car.atom.literal != "NIL" {
			return mkNil(), ErrEval
		}

		if expr.cdr.car.isConsCell() {
			if expr.cdr.car.car.isSexpr() {

				result, _ := expr.cdr.car.Eval()

				if result.atom.num.Cmp(big.NewInt(0)) == 0 {

					return mkSymbolTrue(), nil
				} else {
					return mkNil(), nil
				}
			}
		} else {

			if expr.cdr.car.atom.num.Cmp(big.NewInt(0)) == 0 {

				return mkSymbolTrue(), nil
			} else {

				return mkNil(), nil
			}
		}

	}

	return nil, ErrEval
}

func (expr *SExpr) isSexpr() bool {
	if expr.atom.literal == "QUOTE" {
		return true
	} else if expr.atom.literal == "CAR" {
		return true
	} else if expr.atom.literal == "CDR" {
		return true
	} else if expr.atom.literal == "CONS" {
		return true
	} else if expr.atom.literal == "+" {
		return true
	} else if expr.atom.literal == "*" {
		return true
	} else if expr.atom.literal == "LENGTH" {
		return true
	}
	return false
}

func (expr *SExpr) quote() (*SExpr, error) {

	if !expr.cdr.cdr.isNil() && expr.cdr.car.isAtom() && expr.cdr.car.atom.literal != "NIL" {
		return mkNil(), ErrEval
	}

	return expr.cdr.car, nil
}

func (expr *SExpr) cons() (*SExpr, error) {

	if expr.cdr.cdr.car.isConsCell() {
		if expr.cdr.cdr.car.car.isSexpr() {
			temp := expr.cdr.cdr.car
			res, _ := temp.Eval()
			return mkConsCell(expr.cdr.car, res), nil
		}
	}

	if expr.cdr.car.atom.literal == "NIL" && expr.cdr.cdr.car.atom.literal == "NIL" {
		return mkConsCell(mkNil(), mkNil()), nil
	}
	if expr.cdr.cdr.cdr.car != nil {
		return mkNil(), ErrEval
	}
	if expr.cdr.car.isSymbol() && expr.cdr.car.atom.literal != "NIL" {
		return mkNil(), ErrEval
	}
	if expr.cdr.cdr.car.isSymbol() {
		return mkNil(), ErrEval
	}
	return mkConsCell(expr.cdr.car, expr.cdr.cdr.car), nil
}

func (expr *SExpr) carr() (*SExpr, error) {
	if expr.cdr.cdr.car != nil {
		return nil, ErrEval
	}

	if expr.cdr.car.isConsCell() {

		if expr.cdr.car.car.isSexpr() {
			ex, _ := expr.cdr.car.Eval()
			return ex.car, nil
		}

	}

	if expr.cdr.car.atom.literal == "NIL" {
		return expr.cdr.car, nil
	}

	return nil, ErrEval
}

func (expr *SExpr) cdrr() (*SExpr, error) {
	if expr.cdr.car.isConsCell() {
		if expr.cdr.car.car.isSexpr() {
			ex, _ := expr.cdr.car.Eval()
			return ex.cdr, nil
		}

	}

	if expr.cdr.car.atom.literal == "NIL" {
		return expr.cdr.car, nil
	}

	return nil, ErrEval
}

func (expr *SExpr) add() (*SExpr, error) {

	var t *SExpr = new(SExpr)
	t = expr.cdr

	sum := big.NewInt(0)

	for t.car != nil {

		if t.car.isConsCell() {

			if t.car.car.isSexpr() {

				temp, _ := t.car.Eval()
				sum.Add(sum, temp.atom.num)
				t = t.cdr
			}
		} else {
			sum.Add(sum, t.car.atom.num)
			t = t.cdr

		}
	}
	return mkNumber(sum), nil
}

func (expr *SExpr) mult() (*SExpr, error) {
	var t *SExpr = new(SExpr)
	t = expr.cdr
	sum := big.NewInt(1)

	for t.car != nil {

		if t.car.isConsCell() {

			if t.car.car.isSexpr() {

				temp, _ := t.car.Eval()
				sum.Mul(sum, temp.atom.num)
				t = t.cdr
			}
		} else {
			sum.Mul(sum, t.car.atom.num)
			t = t.cdr

		}
	}
	return mkNumber(sum), nil
}

func (expr *SExpr) length(len int64) int64 {

	if expr.cdr.isAtom() {
		return len + 1
	}

	len = expr.cdr.length(len)
	len = len + 1
	return len
}

func (expr *SExpr) list() (*SExpr, error) {
	if expr.cdr.car == nil {
		return nil, ErrEval
	}

	if expr.cdr.car.isConsCell() {
		if expr.cdr.car.car == nil {
			return mkSymbolTrue(), nil
		}

		if expr.cdr.car.car.isSexpr() {
			temp, _ := expr.cdr.car.Eval()
			if expr.cdr.car.car.atom.literal == "CAR" {
				if temp.atom.literal == "NIL" {
					return mkSymbolTrue(), nil
				}
			} else if temp.isConsCell() {
				return mkSymbolTrue(), nil
			}
			return mkNil(), nil
		}
	} else if expr.cdr.car.atom.literal == "NIL" {
		return mkSymbolTrue(), nil
	} else if expr.cdr.car.isNumber() && expr.cdr.cdr.isNil() {
		return mkNil(), nil
	}

	return nil, ErrEval
}
