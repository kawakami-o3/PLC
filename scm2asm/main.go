package main

import (
	"flag"
	"fmt"
	"strconv"
)

func emit(s string, v ...interface{}) {
	fmt.Println(fmt.Sprintf(s, v...))
}

const (
	fixnumShift = 2
	fixnumTag   = 0x00
	charShift   = 8
	charTag     = 0x0F
	boolShift   = 7
	boolTag     = 0x1F
	boolTrue    = 1<<boolShift + boolTag
	boolFalse   = 0<<boolShift + boolTag
	emptyList   = 0x2F
)

func immediateRep(x string) int {
	i, err := strconv.Atoi(x)
	if err == nil {
		return i << fixnumShift
	}

	switch x {
	case "()":
		return emptyList
	case "#t":
		return boolTrue
	case "#f":
		return boolFalse
	default:
		return int(x[0])<<charShift + charTag
	}
}

func isImmediate(e Expr) bool {
	return e.value != ""
}

func isPrimcall(e Expr) bool {
	op := primcallOp(e).value
	for _, s := range primcallOpList {
		if op == s {
			return true
		}
	}
	return false
}

func primcallOp(e Expr) Expr {
	return e.list[0]
}

func primcallOperand1(e Expr) Expr {
	return e.list[1]
}

var primcallOpList = []string{
	"add1",
	"sub1",
	"integer->char",
	"char->integer",
	"zero?",
	"null?",
	"not",
	"integer?",
	"boolean?",
}

func emitComp(target int) {
	emit("\tcmpl $%d, %%eax", target)
	emit("\tmovl $0, %%eax")
	emit("\tsete %%al")
	emit("\tsall $%d, %%eax", boolShift)
	emit("\torl $%d, %%eax", boolTag)
}

func emitExpr(expr Expr) {
	if isImmediate(expr) {
		emit("\tmovl $%d, %%eax", immediateRep(expr.value))
	} else if isPrimcall(expr) {
		emitExpr(primcallOperand1(expr))

		switch primcallOp(expr).value {
		case "add1":
			emit("\taddl $%d, %%eax", immediateRep("1"))
		case "sub1":
			emit("\tsubl $%d, %%eax", immediateRep("1"))
		case "integer->char":
			emit("\tsall $%d, %%eax", charShift-fixnumShift)
			emit("\torl $%d, %%eax", charTag)
		case "char->integer":
			emit("\tsarl $%d, %%eax", charShift-fixnumShift)
		case "zero?":
			emitComp(0)
		case "null?":
			emitComp(emptyList)
		case "not":
			emitComp(boolFalse)
		case "integer?":
			emit("\tandl $%d, %%eax", 1<<fixnumShift-1)
			emitComp(fixnumTag)
		case "boolean?":
			emit("\tandl $%d, %%eax", 1<<boolShift-1)
			emitComp(boolTag)
		}

	} else {
		//
	}
}

type Expr struct {
	value string
	list  []Expr
}

type tokenBuffer struct {
	tokens []string
	idx    int
}

func (this *tokenBuffer) get() string {
	return this.tokens[this.idx]
}

func (this *tokenBuffer) consume() {
	this.idx++
}

func (this *tokenBuffer) next() string {
	this.consume()
	return this.get()
}

func (this *tokenBuffer) hasNext() bool {
	return this.idx+1 < len(this.tokens)
}

func tokenize(x string) *tokenBuffer {
	x += " " // sentinel

	tokens := []string{}
	t := ""
	for _, c := range x {
		if c == ' ' || c == '\n' || c == '\t' {
			if len(t) > 0 {
				tokens = append(tokens, t)
				t = ""
			}
		} else if c == ')' {
			if len(t) == 0 {
				// TODO quote
				if tokens[len(tokens)-1] == "(" {
					tokens[len(tokens)-1] += ")"
				} else {
					tokens = append(tokens, string(c))
				}
			} else {
				tokens = append(tokens, t)
				t = ""
				tokens = append(tokens, string(c))
			}
		} else if c == '(' || c == ')' {
			if len(t) > 0 {
				tokens = append(tokens, t)
				t = ""
			}

			tokens = append(tokens, string(c))
		} else {
			t += string(c)
		}
	}

	return &tokenBuffer{tokens, -1}
}

func makeExpr(tokens *tokenBuffer) Expr {
	t := tokens.get()
	if t == "(" {
		ret := Expr{}
		for tokens.next() != ")" {
			expr := makeExpr(tokens)
			ret.list = append(ret.list, expr)
		}
		tokens.consume()

		if len(ret.list) == 0 {
			ret.value = "()"
		}
		return ret
	} else if t == ")" {
		panic("unexpected ')'")
	} else {
		ret := Expr{}
		ret.value = t
		return ret
	}
}

func parse(x string) Expr {
	tokens := tokenize(x)
	exprs := []Expr{}
	for tokens.hasNext() {
		tokens.next()
		e := makeExpr(tokens)
		exprs = append(exprs, e)
	}

	if len(exprs) == 1 {
		return exprs[0]
	} else {
		expr := Expr{}
		expr.list = exprs
		return expr
	}
}

func compileProgram(x string) {
	emit("\t.text")
	emit("\t.p2align 4,,15")
	emit("\t.globl scheme_entry")
	emit("\t.type scheme_entry, @function")
	emit("scheme_entry:")
	emitExpr(parse(x))
	emit("\tret")
}

func main() {
	flag.Parse()
	target := flag.Args()[0]

	compileProgram(target)
}
