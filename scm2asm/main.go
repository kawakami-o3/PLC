package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"unicode"
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
func isBool(e expression) bool {
	return e.value == tokenTrue || e.value == tokenFalse
}
*/

func isImmediate(e expression) bool {
	x := e.value
	if len(x) == 0 {
		return false
	}

	_, err := immediateRep(e.value)
	return err == nil
}

var (
	specialInitial = map[byte]int{
		'!': 0,
		'$': 0,
		'%': 0,
		'&': 0,
		'*': 0,
		'/': 0,
		':': 0,
		'<': 0,
		'=': 0,
		'>': 0,
		'?': 0,
		'^': 0,
		'_': 0,
		'~': 0,
	}

	specialSubsequent = map[byte]int{
		'+': 0,
		'-': 0,
		'.': 0,
		'@': 0,
	}

	syntacticKeyword = map[string]int{
		"else":             0,
		"=>":               0,
		"define":           0,
		"unquote":          0,
		"unquote-splicing": 0,
		"quote":            0,
		"lambda":           0,
		"if":               0,
		"set!":             0,
		"begin":            0,
		"cond":             0,
		"and":              0,
		"or":               0,
		"case":             0,
		"let":              0,
		"let*":             0,
		"letrec":           0,
		"do":               0,
		"delay":            0,
		"quasiquote":       0,
	}

	peculiarIdentifier = map[string]int{
		"+":   0,
		"-":   0,
		"...": 0,
	}
)

func isInitial(c byte) bool {
	_, ok := specialInitial[c]
	return ok || unicode.IsLetter(rune(c))
}

func isSubsequent(c byte) bool {
	_, ok := specialSubsequent[c]
	return isInitial(c) || unicode.IsDigit(rune(c)) || ok
}

func isVariable(e expression) bool {
	x := e.value
	if len(x) == 0 {
		return false
	}

	_, ng := syntacticKeyword[x]
	if ng {
		return false
	}

	_, piOk := peculiarIdentifier[x]
	if piOk {
		return true
	}

	if !isInitial(e.value[0]) {
		return false
	}
	for _, c := range e.value[1:] {
		if !isSubsequent(byte(c)) {
			return false
		}
	}
	return true
}

func isLet(e expression) bool {
	return len(e.list) > 0 && e.list[0].value == "let"
}

func isPrimcall(e expression) bool {
	op := primcallOp(e).value
	for _, s := range primcallOpList {
		if op == s {
			return true
		}
	}
	return false
}

func primcallOp(e expression) expression {
	return e.list[0]
}

func primcallOperand1(e expression) expression {
	return e.list[1]
}

func primcallOperand2(e expression) expression {
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

func emitOperand2(expr expression, si int, env *environment) {
	emitExpr(primcallOperand2(expr), si, env)
	emit("\tmovl %%eax, %d(%%rsp)", si)
	emitExpr(primcallOperand1(expr), si-wordSize, env)
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

func emitPrimitiveCall(expr expression, si int, env *environment) {
	switch primcallOp(expr).value {
	case "add1":
		emitExpr(primcallOperand1(expr), si, env)
		emit("\taddl $%d, %%eax", fixnum1)
	case "sub1":
		emitExpr(primcallOperand1(expr), si, env)
		emit("\tsubl $%d, %%eax", fixnum1)
	case "integer->char":
		emitExpr(primcallOperand1(expr), si, env)
		emit("\tsall $%d, %%eax", charShift-fixnumShift)
		emit("\torl $%d, %%eax", charTag)
	case "char->integer":
		emitExpr(primcallOperand1(expr), si, env)
		emit("\tsarl $%d, %%eax", charShift-fixnumShift)
	case "zero?":
		emitExpr(primcallOperand1(expr), si, env)
		emitEq(0)
	case "null?":
		emitExpr(primcallOperand1(expr), si, env)
		emitEq(emptyList)
	case "not":
		emitExpr(primcallOperand1(expr), si, env)
		emitEq(boolFalse)
	case "integer?":
		emitExpr(primcallOperand1(expr), si, env)
		emit("\tandl $%d, %%eax", 1<<fixnumShift-1)
		emitEq(fixnumTag)
	case "boolean?":
		emitExpr(primcallOperand1(expr), si, env)
		emit("\tandl $%d, %%eax", 1<<boolShift-1)
		emitEq(boolTag)
	case "+":
		emitOperand2(expr, si, env)
		emit("\taddl %d(%%rsp), %%eax", si)
	case "-":
		emitOperand2(expr, si, env)
		emit("\tsubl %d(%%rsp), %%eax", si)
	case "*":
		emitOperand2(expr, si, env)
		emit("\timull %d(%%rsp), %%eax", si)
		emit("sarl $%d, %%eax", fixnumShift)
	case "=":
		emitOperand2(expr, si, env)
		emitCompStack("sete", si)
	case "<":
		emitOperand2(expr, si, env)
		emitCompStack("setl", si)
	case "<=":
		emitOperand2(expr, si, env)
		emitCompStack("setle", si)
	case ">":
		emitOperand2(expr, si, env)
		emitCompStack("setg", si)
	case ">=":
		emitOperand2(expr, si, env)
		emitCompStack("setge", si)
	case "char=?":
		emitOperand2(expr, si, env)
		emitCompStack("sete", si)
	}
}

type environment struct {
	variables map[string]int
}

func newEnv() *environment {
	env := &environment{}
	env.variables = map[string]int{}
	return env
}

func (env *environment) lookup(x expression) (int, error) {
	si, ok := env.variables[x.value]
	if ok {
		return si, nil
	} else {
		return 0, errors.New(fmt.Sprintf("variable not found: %s", x.value))
	}
}

func (env *environment) extend(e expression, si int) {
	env.variables[e.value] = si
}

func lhs(e expression) expression {
	return e.list[0]
}

func rhs(e expression) expression {
	return e.list[1]
}

func bindings(e expression) []expression {
	// named let
	if len(e.list[1].value) > 0 {
		return e.list[2].list
	} else {
		return e.list[1].list
	}
}

func body(e expression) expression {
	i := 2
	if len(e.list[1].value) > 0 {
		i += 1
	}

	if i+1 == len(e.list) {
		return e.list[i]
	} else {
		ret := expression{}
		ret.list = e.list[i:]
		return ret
	}
}

func emitLet(bindings []expression, body expression, si int, env *environment) {
	if len(bindings) == 0 {
		emitExpr(body, si, env)
	} else {
		b := bindings[0]
		emitExpr(rhs(b), si, env)
		emit("\tmovl %%eax, %d(%%rsp)", si)
		env.extend(lhs(b), si)
		emitLet(bindings[1:], body, si-wordSize, env)
	}
}

func emitExpr(expr expression, si int, env *environment) {
	if isImmediate(expr) {
		n, err := immediateRep(expr.value)
		if err != nil {
			panic(err)
		}
		emit("\tmovl $%d, %%eax", n)
	} else if isVariable(expr) {
		n, err := env.lookup(expr)
		if err != nil {
			panic(err)
		}
		emit("\tmovl %d(%%rsp), %%eax", n)
	} else if isLet(expr) {
		emitLet(bindings(expr), body(expr), si, env)
	} else if isPrimcall(expr) {
		emitPrimitiveCall(expr, si, env)
	} else {
		//
	}
}

type expression struct {
	value string
	list  []expression
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

func makeExpr(tokens *tokenBuffer) expression {
	t := tokens.get()
	if t == "(" {
		ret := expression{}
		for tokens.next() != ")" {
			expr := makeExpr(tokens)
			ret.list = append(ret.list, expr)
		}

		if len(ret.list) == 0 {
			ret.value = "()"
		}
		return ret
	} else if t == ")" {
		panic("unexpected ')'")
	} else {
		ret := expression{}
		ret.value = t
		return ret
	}
}

func parse(x string) expression {
	tokens := tokenize(x)
	exprs := []expression{}
	for tokens.hasNext() {
		tokens.next()
		e := makeExpr(tokens)
		exprs = append(exprs, e)
	}

	if len(exprs) == 1 {
		return exprs[0]
	} else {
		expr := expression{}
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
	emitExpr(parse(x), stackIndexInit, newEnv())
	emit("\tret")
}

func main() {
	flag.Parse()
	target := flag.Args()[0]

	compileProgram(target)
}
