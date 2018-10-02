package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"unicode"
)

const (
	bytes = 4
	//byteSize = 8

	//fixnumBits  = wordSize*byteSize - fixnumShift
	fixnumBits  = 32 - fixnumShift
	fixnumLower = -(1 << fixnumBits)
	fixnumUpper = 1<<fixnumBits - 1
	fixnumShift = 2
	fixnumTag   = 0x00
	fixnum1     = 1 << fixnumShift

	boolShift = 7
	boolTag   = 0x1F
	boolTrue  = 1<<boolShift + boolTag
	boolFalse = 0<<boolShift + boolTag

	emptyList = 0x2F

	charShift = 8
	charTag   = 0x0F

	objShift  = 3
	objMask   = 0x07
	pairTag   = 0x01
	pairSize  = 16
	pairCar   = 0
	pairCdr   = 8
	vectorTag = 0x05
	stringTag = 0x06

	wordSize       = 8
	stackIndexInit = -wordSize

	tokenTrue  = "#t"
	tokenFalse = "#f"
	tokenEmpty = "()"

	sete = "sete"
	setx = "setx"

	heapCellSize = 8
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
		"delay":            0, "quasiquote": 0,
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

func isEmpty(e expression) bool {
	return e.value == tokenEmpty
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

func isIf(e expression) bool {
	if len(e.list) == 0 {
		return false
	}

	return e.list[0].value == "if"
}

func isLet(e expression) bool {
	if len(e.list) == 0 {
		return false
	}

	v := e.list[0].value
	return v == "let" || v == "let*"
}

func isLetrec(e expression) bool {
	if len(e.list) == 0 {
		return false
	}

	v := e.list[0].value
	return v == "letrec"
}

func isCarCdr(op string) bool {
	if len(op) >= 3 && op[0] == 'c' && op[len(op)-1] == 'r' {
		for i := 1; i < len(op)-1; i++ {
			c := op[i]
			if c != 'a' && c != 'd' {
				return false
			}
		}
		return true
	}
	return false
}

func isPrimcall(e expression) bool {
	op := primcallOp(e).value
	for _, s := range primcallOpList {
		if op == s {
			return true
		}
	}
	return isCarCdr(op)
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

func emit(s string, v ...interface{}) {
	fmt.Println(fmt.Sprintf(s, v...))
}

func addr(name string, index int) string {
	if index == 0 {
		return name
	} else {
		return fmt.Sprintf("%d(%s)", index, name)
	}
}

func eax(index int) string {
	return addr("%eax", index)
}

func rsp(index int) string {
	return addr("%rsp", index)
}

func esi(index int) string {
	//return fmt.Sprintf("%d(%%esi)", index)
	return addr("%esi", index)
}

func num(i int) string {
	return fmt.Sprintf("$%d", i)
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

func emitHeapAlloc(size int) {
	allocSize := (((size - 1) / heapCellSize) + 1) * heapCellSize
	emit("\tmov %%rbp, %%rax")
	emit("\tadd $%d, %%rbp", allocSize*bytes)
}

func emitHeapAllocDynamic() {
	emit("\tsub $1, %%rax")
	emit("\tshr $%d, %%rax", objShift)
	emit("\tadd $1, %%rax")
	emit("\tshl $%d, %%rax", objShift+2)
	emit("\tmov %%rbp, %%rdx")
	emit("\tadd %%rbp, %%rbp")
	emit("\tmov %%rdx, %%rax")
}

func emitStackLoad(si int) {
	emitMov(rsp(si), eax(0))
}

func emitStackSave(si int) {
	//emitMov(eax(0), rsp(si))
	emit("\tmov %%rax, %d(%%rsp)", si)
}

func emitStackToHeap(si, offset int) {
	emit("\tmov %d(%%rsp), %%rdx", si)
	emit("\tmov %%rdx, %d(%%rax)", offset)
}

func emitCmpBool(cmp string) {
	// cmp: sete, setx ,,,
	emit("\t%s %%al", cmp)
	emit("\tmovzbq %%al, %%rax")
	emit("\tsal $%d, %%al", boolShift)
	emit("\tor $%d, %%al", boolFalse)
}

func emitIsObject(expr expression, si int, env *environment, tag int) {
	emitExpr(expr, si, env)
	emit("\tand $%d, %%al", objMask)
	emit("\tcmp $%d, %%al", tag)
	emitCmpBool(sete)
}

func emitOperand2(expr expression, si int, env *environment) {
	emitExpr(primcallOperand1(expr), si-wordSize, env)
	emitStackSave(si)
	emitExpr(primcallOperand2(expr), si, env)
}

func emitMov(a, b string) {
	emit("\tmovl %s, %s", a, b)
}

func emitOrl(a, b string) {
	emit("\torl %s, %s", a, b)
}

func emitAdd(a, b string) {
	emit("\taddl %s, %s", a, b)
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

	"cons",
	"pair?",

	"make-vector",
	"vector?",

	"make-string",
	"string?",
}

func nextStackIndex(si int) int {
	return si - wordSize
}

func emitPrimitiveCall(expr expression, si int, env *environment) {
	op := primcallOp(expr).value
	switch op {
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
		emit("\tsarl $%d, %%eax", fixnumShift)
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
	case "car":
		emitExpr(primcallOperand1(expr), si, env)
		//emitMov(eax(-1), eax(0))
		n, _ := immediateRep("10")
		emitMov(num(n), eax(0))
	case "cons":
		emitOperand2(expr, si, env)
		emitStackSave(nextStackIndex(si))
		emitHeapAlloc(pairSize)
		emit("\tor $%d, %%rax", pairTag)
		emitStackToHeap(si, pairCar-pairTag)
		emitStackToHeap(nextStackIndex(si), pairCdr-pairTag)
	case "pair?":
		emitIsObject(expr.list[1], si, env, pairTag)

	case "make-vector":
		emitExpr(primcallOperand1(expr), si, env)
		emitStackSave(si)
		emit("\tshr $%d, %%rax", fixnumShift)
		emit("\tadd $1, %%rax")
		emit("\tshl $%d, %%rax", fixnumShift)
		emitHeapAllocDynamic()
		emitStackToHeap(si, 0)
		emit("\tor $%d, %%rax", vectorTag)
	case "vector?":
		emitIsObject(expr.list[1], si, env, vectorTag)

	case "make-string":
		emitExprSave(primcallOperand1(expr), si, env)
		emit("\tshr $%d, %%rax", fixnumShift)
		emit("\tadd $%d, %%rax", wordSize)
		emitHeapAllocDynamic()
		emitStackToHeap(si, 0)
		emit("\tor $%d, %%rax", stringTag)
	case "string?":
		emitIsObject(expr.list[1], si, env, stringTag)
	default:
		if isCarCdr(op) {
		}
	}
}

type environment struct {
	variables map[string]int

	labels map[string]string
}

func newEnv() *environment {
	env := &environment{}
	env.variables = map[string]int{}
	env.labels = map[string]string{}
	return env
}

func makeInitialEnv(lvars []expression, labels []string) *environment {
	env := newEnv()
	for i := 0; i < len(lvars); i++ {
		//env.exps[labels[i]] = lvars[i]
		env.labels[lvars[i].value] = labels[i]
	}
	return env
}

func (env *environment) lookupLabel(x expression) (string, error) {
	label, ok := env.labels[x.value]
	if ok {
		return label, nil
	} else {
		return "", errors.New(fmt.Sprintf("variable not found: %s", x.value))
	}
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

func emitLet(bindings []expression, body expression, si int, env *environment, isTail bool) {
	for _, b := range bindings {
		emitExpr(rhs(b), si, env)
		emit("\tmovl %%eax, %d(%%rsp)", si)
		env.extend(lhs(b), si)

		si -= wordSize
	}

	emitAnyExpr(body, si, env, isTail)
}

func mapLhs(bindings []expression) []expression {
	ret := []expression{}
	for _, b := range bindings {
		ret = append(ret, b.list[0])
	}
	return ret
}

func mapRhs(bindings []expression) []expression {
	ret := []expression{}
	for _, b := range bindings {
		ret = append(ret, b.list[1])
	}
	return ret
}

func letrecBindings(expr expression) []expression {
	return expr.list[1].list
}
func letrecBody(expr expression) expression {
	return expr.list[2]
}

func emitLetrec(expr expression) {
	bindings := letrecBindings(expr)
	lvars := mapLhs(bindings)
	lambdas := mapRhs(bindings)
	labels := uniqueLabels(lvars)
	env := makeInitialEnv(lvars, labels)
	for i := 0; i < len(lambdas); i++ {
		emitLambda(env, lambdas[i], labels[i])
	}

	emitSchemeEntry(letrecBody(expr), env)
}

func lambdaFormals(expr expression) expression {
	return expr.list[1]
}

func lambdaBody(expr expression) expression {
	return expr.list[2]
}

func emitLambda(env *environment, expr expression, label string) {
	emitFunctionHeader(label)
	fmls := lambdaFormals(expr)
	body := lambdaBody(expr)
	si := -wordSize

	for _, e := range fmls.list {
		env.extend(e, si)
		si -= wordSize
	}

	emitTailExpr(body, si, env)
}

func emitSchemeEntry(expr expression, env *environment) {
	emitFunctionHeader("L_scheme_entry")
	emitTailExpr(expr, -wordSize, env)
}

var labelCount = 0

func uniqueLabel() string {
	label := fmt.Sprintf("L_%d", labelCount)
	labelCount++
	return label
}

// TODO need the migration?
func uniqueLabels(lvars []expression) []string {
	ret := []string{}
	for _, lvar := range lvars {
		labelCount++
		ret = append(ret, fmt.Sprintf("L_%s_%d", lvar.value, labelCount))
	}
	return ret
}

func emitJe(label string) {
	emit("\tje %s", label)
}

func emitJmp(label string) {
	emit("\tjmp %s", label)
}

func emitLabel(label string) {
	emit("%s:", label)
}

func emitCmpl(a, b string) {
	emit("\tcmpl %s, %s", a, b)
}

func emitRetIf(isTail bool) {
	if isTail {
		emit("\tret")
	}
}

func emitIf(test, conseq, altern expression, si int, env *environment, isTail bool) {
	altLabel := uniqueLabel()
	endLabel := uniqueLabel()
	emitExpr(test, si, env)
	emitCmpl(num(boolFalse), eax(0))
	emitJe(altLabel)
	emitAnyExpr(conseq, si, env, isTail)
	if !isTail {
		emitJmp(endLabel)
	}
	emitLabel(altLabel)
	emitAnyExpr(altern, si, env, isTail)
	emitLabel(endLabel)
}

func isApp(expr expression, env *environment) bool {
	_, ok := env.labels[expr.list[0].value]
	return ok
}

func emitArguments(args []expression, si int, env *environment) {
	for _, e := range args {
		emitExpr(e, si, env)
		emitStackSave(si)
		si -= wordSize
	}
}

func moveArguments(args []expression, si, delta int, env *environment) {
	if delta == 0 {
		return
	}

	for i := 0; i < len(args); i++ {
		emitStackLoad(si)
		emitStackSave(si + delta)
		si -= wordSize
	}
}

func emitApp(expr expression, si int, env *environment, isTail bool) {
	callTarget := expr.list[0]
	callArgs := expr.list[1:]
	if isTail {
		emitArguments(callArgs, si, env)
		moveArguments(callArgs, si, -(si + wordSize), env)

		label, err := env.lookupLabel(callTarget)
		if err != nil {
			panic(fmt.Sprintf("%s not found", callTarget.value))
		}
		emitJmp(label)
	} else {
		emitArguments(callArgs, si-2*wordSize, env)
		emitAdjustBase(si + wordSize)
		label, err := env.lookupLabel(callTarget)
		if err != nil {
			panic(fmt.Sprintf("%s not found", callTarget.value))
		}
		emitCall(label)
		emitAdjustBase(-1 * (si + wordSize))
	}
}

func emitExprSave(expr expression, si int, env *environment) {
	emitExpr(expr, si, env)
	emitStackSave(si)
}
func emitExpr(expr expression, si int, env *environment) {
	emitAnyExpr(expr, si, env, false)
}

func emitTailExpr(expr expression, si int, env *environment) {
	emitAnyExpr(expr, si, env, true)
}

func emitAnyExpr(expr expression, si int, env *environment, isTail bool) {
	log := ""
	if isImmediate(expr) {
		log += "immediate"
		n, err := immediateRep(expr.value)
		if err != nil {
			panic(err)
		}
		emit("\tmov $%d, %%rax", n)
		emitRetIf(isTail)
	} else if isVariable(expr) {
		log += "variable"
		if n, err := env.lookup(expr); err == nil {
			emitStackLoad(n)
		}
		emitRetIf(isTail)
	} else if isIf(expr) {
		log += "if"
		emitIf(expr.list[1], expr.list[2], expr.list[3], si, env, isTail)
	} else if isLet(expr) {
		log += "let"
		emitLet(bindings(expr), body(expr), si, env, isTail)
	} else if isLetrec(expr) {
		log += "letrec"
		//emitLetrec(bindings(expr), body(expr), si, env)
		emitLetrec(expr)
	} else if isPrimcall(expr) {
		log += "primcall"
		emitPrimitiveCall(expr, si, env)
		emitRetIf(isTail)
	} else if isApp(expr, env) {
		log += "app"
		emitApp(expr, si, env, isTail)
	} else {
		panic(fmt.Sprintf("[emitAnyExpr] not implemented. %v", expr))
	}

	//fmt.Println("[debug]", log)
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
		} else if c == ')' || c == ']' {
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
		} else if c == '(' || c == ')' || c == '[' || c == ']' {
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

func isBra(s string) bool {
	return s == "(" || s == "["
}

func isKet(s string) bool {
	return s == ")" || s == "]"
}

func makeExpr(tokens *tokenBuffer) expression {
	t := tokens.get()
	if isBra(t) {
		ret := expression{}
		for false == isKet(tokens.next()) {
			expr := makeExpr(tokens)
			ret.list = append(ret.list, expr)
		}

		if len(ret.list) == 0 {
			ret.value = "()"
		}
		return ret
	} else if isKet(t) {
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

func emitFunctionHeader(label string) {
	emit("\t.text")
	emit("\t.global %s", label)
	emit("\t.type %s, @function", label)
	emitLabel(label)
}

func emitAdjustBase(si int) {
	if si != 0 {
		emit("\taddq $%d, %%rsp", si)
	}
}

func emitCall(label string) {
	emit("\tcall %s", label)
}

func compileProgram(x string) {
	//emitFunctionHeader("L_scheme_entry")
	//emitExpr(parse(x), stackIndexInit, newEnv())
	//emit("\tret")

	emitFunctionHeader("scheme_entry")

	emit("\tmov %%rdi, %%rcx")
	emit("\tmov %%rbx, 8(%%rcx)")
	emit("\tmov %%rsi, 32(%%rcx)")
	emit("\tmov %%rdi, 40(%%rcx)")
	emit("\tmov %%rbp, 48(%%rcx)")
	emit("\tmov %%rsp, 56(%%rcx)")
	emit("\tmov %%rdx, %%rbp")
	emit("\tmov %%rsi, %%rsp")
	emit("\tcall L_scheme_entry")
	emit("\tmov 8(%%rcx), %%rbx")
	emit("\tmov 32(%%rcx), %%rsi")
	emit("\tmov 40(%%rcx), %%rdi")
	emit("\tmov 48(%%rcx), %%rbp")
	emit("\tmov 56(%%rcx), %%rsp")
	emit("\tret")

	program := parse(x)
	if isLetrec(program) {
		emitLetrec(program)
	} else {
		emitSchemeEntry(program, makeInitialEnv([]expression{}, []string{}))
	}
}

func main() {
	flag.Parse()
	target := flag.Args()[0]

	compileProgram(target)
}
