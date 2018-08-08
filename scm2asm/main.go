package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
)

func emit(s string, v ...interface{}) {
	fmt.Println(fmt.Sprintf(s, v...))
}

const (
	byteSize = 8

	fixnumBits  = wordSize*byteSize - fixnumShift
	fixnumLower = -(1 << fixnumBits)
	fixnumUpper = 1<<fixnumBits - 1
	fixnumShift = 2
	fixnumTag   = 0x00
	fixnum1     = 1 << fixnumShift

	charShift = 8
	charTag   = 0x0F
	boolShift = 7
	boolTag   = 0x1F
	boolTrue  = 1<<boolShift + boolTag
	boolFalse = 0<<boolShift + boolTag
	emptyList = 0x2F

	stackIndexInit = -4
	wordSize       = 4

	tokenTrue  = "#t"
	tokenFalse = "#f"
	tokenEmpty = "()"
)

func immediateRep(x string) (int, error) {
	i, err := strconv.Atoi(x)
	if err == nil {
		return i << fixnumShift, nil
	}

	// char
	if len(x) >= 3 && x[0:2] == "#\\" {
		if x == "#\\space" {
			return int(' ')<<charShift + charTag, nil
		} else if x == "#\\newline" {
			return int('\n')<<charShift + charTag, nil
		} else {
			return int(x[2])<<charShift + charTag, nil
		}
	}

	switch x {
	case tokenEmpty:
		return emptyList, nil
	case tokenTrue:
		return boolTrue, nil
	case tokenFalse:
		return boolFalse, nil
	}
	return -1, errors.New(fmt.Sprintf("not an immediate, %s", x))
}

/*
// TODO use in isImmediate
func isBool(e Expr) bool {
	return e.value == tokenTrue || e.value == tokenFalse
}
*/

func isImmediate(e Expr) bool {
	x := e.value
	if len(x) == 0 {
		return false
	}

	_, err := immediateRep(e.value)
	return err == nil
}

func isVariable(e Expr) bool {
	// TODO
	return false
}

func isLet(e Expr) bool {
	// TODO
	return false
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

func primcallOperand2(e Expr) Expr {
	return e.list[2]
}

func emitEq(target int) {
	emit("\tcmpl $%d, %%eax", target)
	emit("\tmovl $0, %%eax")
	emit("\tsete %%al")
	emit("\tsall $%d, %%eax", boolShift)
	emit("\torl $%d, %%eax", boolTag)
}

func emitCompStack(op string, si int) {
	emit("\tcmpl %d(%%rsp), %%eax", si)
	emit("\tmovl $0, %%eax")
	emit("\t%s %%al", op)
	emit("\tsall $%d, %%eax", boolShift)
	emit("\torl $%d, %%eax", boolTag)
}

func emitOperand2(expr Expr, si int) {
	emitExpr(primcallOperand2(expr), si)
	emit("\tmovl %%eax, %d(%%rsp)", si)
	emitExpr(primcallOperand1(expr), si-wordSize)
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
	"+",
	"-",
	"*",
	"=",
	"<",
	"<=",
	">",
	">=",
	"char=?",
}

func emitPrimitiveCall(expr Expr, si int) {
	switch primcallOp(expr).value {
	case "add1":
		emitExpr(primcallOperand1(expr), si)
		emit("\taddl $%d, %%eax", fixnum1)
	case "sub1":
		emitExpr(primcallOperand1(expr), si)
		emit("\tsubl $%d, %%eax", fixnum1)
	case "integer->char":
		emitExpr(primcallOperand1(expr), si)
		emit("\tsall $%d, %%eax", charShift-fixnumShift)
		emit("\torl $%d, %%eax", charTag)
	case "char->integer":
		emitExpr(primcallOperand1(expr), si)
		emit("\tsarl $%d, %%eax", charShift-fixnumShift)
	case "zero?":
		emitExpr(primcallOperand1(expr), si)
		emitEq(0)
	case "null?":
		emitExpr(primcallOperand1(expr), si)
		emitEq(emptyList)
	case "not":
		emitExpr(primcallOperand1(expr), si)
		emitEq(boolFalse)
	case "integer?":
		emitExpr(primcallOperand1(expr), si)
		emit("\tandl $%d, %%eax", 1<<fixnumShift-1)
		emitEq(fixnumTag)
	case "boolean?":
		emitExpr(primcallOperand1(expr), si)
		emit("\tandl $%d, %%eax", 1<<boolShift-1)
		emitEq(boolTag)
	case "+":
		emitOperand2(expr, si)
		emit("\taddl %d(%%rsp), %%eax", si)
	case "-":
		emitOperand2(expr, si)
		emit("\tsubl %d(%%rsp), %%eax", si)
	case "*":
		emitOperand2(expr, si)
		emit("\timull %d(%%rsp), %%eax", si)
		emit("sarl $%d, %%eax", fixnumShift)
	case "=":
		emitOperand2(expr, si)
		emitCompStack("sete", si)
	case "<":
		emitOperand2(expr, si)
		emitCompStack("setl", si)
	case "<=":
		emitOperand2(expr, si)
		emitCompStack("setle", si)
	case ">":
		emitOperand2(expr, si)
		emitCompStack("setg", si)
	case ">=":
		emitOperand2(expr, si)
		emitCompStack("setge", si)
	case "char=?":
		emitOperand2(expr, si)
		emitCompStack("sete", si)
	}
}

func emitExpr(expr Expr, si int) {
	if isImmediate(expr) {
		n, err := immediateRep(expr.value)
		if err != nil {
			panic(err)
		}
		emit("\tmovl $%d, %%eax", n)
	} else if isVariable(expr) {
		// TODO
	} else if isLet(expr) {
		// TODO
	} else if isPrimcall(expr) {
		emitPrimitiveCall(expr, si)
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
	emitExpr(parse(x), stackIndexInit)
	emit("\tret")
}

func main() {
	flag.Parse()
	target := flag.Args()[0]

	compileProgram(target)
}
