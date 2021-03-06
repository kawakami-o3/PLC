package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"unicode"

	"github.com/k0kubun/pp"
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

	closureTag = 0x02

	wordSize       = 8
	wordShift      = 3
	stackIndexInit = -wordSize

	tokenTrue  = "#t"
	tokenFalse = "#f"
	tokenEmpty = "()"

	sete  = "sete"
	setl  = "setl"
	setle = "setle"
	setg  = "setg"
	setge = "setge"
	setx  = "setx"

	closure = "closure"

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

func isVariable(e expression) bool {
	// = symbol?
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

func isBegin(e expression) bool {
	if len(e.list) < 2 {
		return false
	}

	return e.list[0].value == "begin"
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

func isList(e expression) bool {
	return len(e.list) > 0
}

func isTaggedList(tag string, e expression) bool {
	// TODO not null?
	return isList(e) && e.list[0].value == tag
}

func isClosure(e expression) bool {
	return isTaggedList(closure, e)
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
	for s, _ := range primcallOpMap {
		if op == s {
			return true
		}
	}
	return isCarCdr(op)
}

func primcallOp(e expression) expression {
	return e.list[0]
}

func operand1(e expression) expression {
	return e.list[1]
}

func operand2(e expression) expression {
	return e.list[2]
}

func operand3(e expression) expression {
	return e.list[3]
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

func rax(index int) string {
	return addr("%rax", index)
}

func rsp(index int) string {
	return addr("%rsp", index)
}

/*
func esi(index int) string {
	//return fmt.Sprintf("%d(%%esi)", index)
	return addr("%esi", index)
}

func num(i int) string {
	return fmt.Sprintf("$%d", i)
}
*/

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
	emit("\tadd %%rax, %%rbp")
	emit("\tmov %%rdx, %%rax")
}

func emitHeapLoad(offset int) {
	emit("\tmov %d(%%rax), %%rax", offset)
}

func emitStackLoad(si int) {
	emitMov(rsp(si), rax(0))
}

func emitStackSave(si int) {
	//emitMov(rax(0), rsp(si))
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

func emitCmpBinop(cmp string, si int, env *environment, args ...expression) {
	emitBinop(args[0], args[1], si, env)
	emit("\tcmp %%rax, %d(%%rsp)", si)
	emitCmpBool(cmp)
}

func emitIsObject(expr expression, si int, env *environment, tag int) {
	emitExpr(expr, si, env)
	emit("\tand $%d, %%al", objMask)
	emit("\tcmp $%d, %%al", tag)
	emitCmpBool(sete)
}

//func emitBinop(expr expression, si int, env *environment) {
func emitBinop(arg1, arg2 expression, si int, env *environment) {
	//emitExpr(operand1(expr), si, env)
	emitExpr(arg1, si, env)
	emitStackSave(si)
	//emitExpr(operand2(expr), nextStackIndex(si), env)
	emitExpr(arg2, nextStackIndex(si), env)
}

func emitMov(a, b string) {
	emit("\tmov %s, %s", a, b)
}

/*
func emitOrl(a, b string) {
	emit("\torl %s, %s", a, b)
}

func emitAdd(a, b string) {
	emit("\tadd %s, %s", a, b)
}
*/

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

	"car",
	"cdr",
	"cons",
	"pair?",
	"set-car!",
	"set-cdr!",
	"eq?",

	"make-vector",
	"vector?",
	"vector-length",
	"vector-set!",
	"vector-ref",

	"make-string",
	"string?",
	"string-set!",
	"string-ref",
	"string-length",
}

var primcallOpMap map[string]func(expression, int, *environment)

func initPrimCallOpMap() {
	primcallOpMap = map[string]func(expression, int, *environment){}

	primcallOpMap["procedure?"] = emitIsProcedure
}

func emitIsProcedure(expr expression, si int, env *environment) {
	emitIsObject(expr.list[1], si, env, closureTag)
}

func nextStackIndex(si int) int {
	return si - wordSize
}

func emitPrimitiveCall(expr expression, si int, env *environment) {
	op := primcallOp(expr).value
	switch op {
	case "add1":
		emitExpr(operand1(expr), si, env)
		emit("\tadd $%d, %%rax", fixnum1)
	case "sub1":
		emitExpr(operand1(expr), si, env)
		emit("\tsub $%d, %%rax", fixnum1)
	case "integer->char":
		emitExpr(operand1(expr), si, env)
		emit("\tshl $%d, %%rax", charShift-fixnumShift)
		emit("\tor $%d, %%rax", charTag)
	case "char->integer":
		emitExpr(operand1(expr), si, env)
		emit("\tshr $%d, %%rax", charShift-fixnumShift)
	case "zero?":
		emitExpr(operand1(expr), si, env)
		emit("\tcmp $%d, %%rax", fixnumTag)
		emitCmpBool(sete)
	case "null?":
		emitExpr(operand1(expr), si, env)
		emit("\tcmp $%d, %%al", emptyList)
		emitCmpBool(sete)
	case "not":
		emitExpr(operand1(expr), si, env)
		emit("\tcmp $%d, %%al", boolFalse)
		emitCmpBool(sete)
	case "integer?":
		emitExpr(operand1(expr), si, env)
		emit("\tand $%d, %%al", 1<<fixnumShift-1)
		emit("\tcmp $%d, %%al", fixnumTag)
		emitCmpBool(sete)
	case "boolean?":
		emitExpr(operand1(expr), si, env)
		emit("\tand $%d, %%al", 1<<boolShift-1)
		emit("\tcmp $%d, %%al", boolTag)
		emitCmpBool(sete)
	case "+":
		emitBinop(operand1(expr), operand2(expr), si, env)
		emit("\tadd %d(%%rsp), %%rax", si)
	case "-":
		emitBinop(operand1(expr), operand2(expr), si, env)
		emit("\tsub %%rax, %d(%%rsp)", si)
		emitStackLoad(si)
	case "*":
		emitBinop(operand1(expr), operand2(expr), si, env)
		emit("\tshr $%d, %%rax", fixnumShift)
		emit("\tmulq %d(%%rsp)", si)
	case "=":
		emitCmpBinop(sete, si, env, expr.list...)
	case "<":
		emitCmpBinop(setl, si, env, expr.list...)
	case "<=":
		emitCmpBinop(setle, si, env, expr.list...)
	case ">":
		emitCmpBinop(setg, si, env, expr.list...)
	case ">=":
		emitCmpBinop(setge, si, env, expr.list...)
	case "char=?":
		emitCmpBinop(sete, si, env, expr.list...)
	case "car":
		emitExpr(operand1(expr), si, env)
		emitHeapLoad(pairCar - pairTag)
	case "cdr":
		emitExpr(operand1(expr), si, env)
		emitHeapLoad(pairCdr - pairTag)
	case "cons":
		emitBinop(operand1(expr), operand2(expr), si, env)
		emitStackSave(nextStackIndex(si))
		emitHeapAlloc(pairSize)
		emit("\tor $%d, %%rax", pairTag)
		emitStackToHeap(si, pairCar-pairTag)
		emitStackToHeap(nextStackIndex(si), pairCdr-pairTag)
	case "pair?":
		emitIsObject(expr.list[1], si, env, pairTag)
	case "set-car!":
		emitBinop(operand2(expr), operand1(expr), si, env)
		emitStackToHeap(si, pairCar-pairTag)
	case "set-cdr!":
		emitBinop(operand2(expr), operand1(expr), si, env)
		emitStackToHeap(si, pairCdr-pairTag)
	case "eq?":
		emitBinop(operand1(expr), operand2(expr), si, env)
		emit("\tcmp %d(%%rsp), %%rax", si)
		emitCmpBool(sete)

	case "make-vector":
		emitExpr(operand1(expr), si, env)
		emitStackSave(si)
		emit("\tshr $%d, %%rax", fixnumShift)
		emit("\tadd $1, %%rax")
		emit("\tshl $%d, %%rax", wordShift)
		emitHeapAllocDynamic()
		emitStackToHeap(si, 0)
		emit("\tor $%d, %%rax", vectorTag)
	case "vector?":
		emitIsObject(expr.list[1], si, env, vectorTag)
	case "vector-length":
		emitExpr(operand1(expr), si, env)
		emitHeapLoad(-vectorTag)
	case "vector-set!":
		vector := operand1(expr)
		index := operand2(expr)
		value := operand3(expr)
		emitExpr(index, si, env)
		emit("\tshl $%d, %%rax", objShift-fixnumShift)
		emit("\tadd $%d, %%rax", wordSize)
		emitStackSave(si)
		emitExprSave(value, nextStackIndex(si), env)
		emitExpr(vector, si, env)
		emit("\tadd %d(%%rsp), %%rax", si)
		emitStackToHeap(nextStackIndex(si), -vectorTag)
	case "vector-ref":
		vector := operand1(expr)
		index := operand2(expr)
		emitExpr(index, si, env)
		emit("\tshl $%d, %%rax", objShift-fixnumShift)
		emit("\tadd $%d, %%rax", wordSize)
		emitStackSave(si)
		emitExpr(vector, si, env)
		emit("\tadd %d(%%rsp), %%rax", si)
		emitHeapLoad(-vectorTag)

	case "make-string":
		emitExprSave(operand1(expr), si, env)
		emit("\tshr $%d, %%rax", fixnumShift)
		emit("\tadd $%d, %%rax", wordSize)
		emitHeapAllocDynamic()
		emitStackToHeap(si, 0)
		emit("\tor $%d, %%rax", stringTag)
	case "string?":
		emitIsObject(expr.list[1], si, env, stringTag)
	case "string-set!":
		str := operand1(expr)
		index := operand2(expr)
		value := operand3(expr)
		emitExpr(index, si, env)
		emit("\tshr $%d, %%rax", fixnumShift)
		emit("\tadd $%d, %%rax", wordSize)
		emitStackSave(si)
		emitExpr(value, nextStackIndex(si), env)
		emit("\tshr $%d, %%rax", charShift)
		emitStackSave(nextStackIndex(si))
		emitExpr(str, si, env)
		emit("\tadd %d(%%rsp), %%rax", si)
		emit("\tmov %d(%%rsp), %%rdx", nextStackIndex(si))
		emit("\tmovb %%dl, %d(%%rax)", -stringTag)
	case "string-ref":
		str := operand1(expr)
		index := operand2(expr)
		emitExpr(index, si, env)
		emit("\tshr $%d, %%rax", fixnumShift)
		emit("\tadd $%d, %%rax", wordSize)
		emitStackSave(si)
		emitExpr(str, si, env)
		emit("\tadd %d(%%rsp), %%rax", si)
		emit("\tmovzbq %d(%%rax), %%rax", -stringTag)
		emit("\tshl $%d, %%rax", charShift)
		emit("\tor $%d, %%rax", charTag)
	case "string-length":
		emitExpr(operand1(expr), si, env)
		emitHeapLoad(-stringTag)

	default:
		if isCarCdr(op) {
		}
	}
	for opLabel, emitter := range primcallOpMap {
		if op == opLabel {
			emitter(expr, si, env)
			return
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
	// for a label
	if len(e.list[1].value) > 0 {
		i += 1
	}

	if i+1 == len(e.list) {
		return e.list[i]
	} else {
		ret := expression{
			list: []expression{expression{value: "begin"}},
		}
		ret.list = append(ret.list, e.list[i:]...)
		return ret
	}
}

func emitLet(bindings []expression, body expression, si int, env *environment, isTail bool) {
	for _, b := range bindings {
		emitExpr(rhs(b), si, env)
		emit("\tmov %%rax, %d(%%rsp)", si)
		env.extend(lhs(b), si)

		si -= wordSize
	}

	emitAnyExpr(si, env, isTail, body)
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
	// FIXME
	return expr.list[2]
	//return body(expr)
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

func emitCmp(a, b string) {
	emit("\tcmp %s, %s", a, b)
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
	emit("\tcmp $%d, %%al", boolFalse)
	emitJe(altLabel)
	emitAnyExpr(si, env, isTail, conseq)
	if !isTail {
		emitJmp(endLabel)
	}
	emitLabel(altLabel)
	emitAnyExpr(si, env, isTail, altern)
	emitLabel(endLabel)
}

func emitBegin(expr expression, si int, env *environment, isTail bool) {
	emitSeq(expr.list[1:], si, env, isTail)
}

func emitSeq(exprs []expression, si int, env *environment, isTail bool) {
	for i, e := range exprs {
		if i == len(exprs)-1 {
			emitAnyExpr(si, env, isTail, e)
		} else {
			emitExpr(e, si, env)
		}
	}
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
		emitArguments(callArgs, si-wordSize, env)
		emitAdjustBase(si + wordSize)
		label, err := env.lookupLabel(callTarget)
		if err != nil {
			panic(fmt.Sprintf("%s not found", callTarget.value))
		}
		emitCall(label)
		emitAdjustBase(-1 * (si + wordSize))
	}
}

func emitClosure(expr expression, si int, env *environment, isTail bool) {
	panic("closure is not implemented yet.")
}

func emitExprSave(expr expression, si int, env *environment) {
	emitExpr(expr, si, env)
	emitStackSave(si)
}

func emitExpr(expr expression, si int, env *environment) {
	emitAnyExpr(si, env, false, expr)
}

func emitTailExpr(expr expression, si int, env *environment) {
	emitAnyExpr(si, env, true, expr)
}

func emitImmediate(expr expression) {
	n, err := immediateRep(expr.value)
	if err != nil {
		panic(err)
	}
	emit("\tmov $%d, %%rax", n)
}

func emitVariableRef(expr expression, si int, env *environment) {
	if n, err := env.lookup(expr); err == nil {
		emitStackLoad(n)
	}
}

func emitAnyExpr(si int, env *environment, isTail bool, expr expression) {
	if isImmediate(expr) {
		emitImmediate(expr)
		emitRetIf(isTail)
	} else if isVariable(expr) {
		emitVariableRef(expr, si, env)
		emitRetIf(isTail)
	} else if isClosure(expr) {
		emitClosure(expr, si, env, isTail)
		emitRetIf(isTail)
	} else if isIf(expr) {
		emitIf(expr.list[1], expr.list[2], expr.list[3], si, env, isTail)
	} else if isLet(expr) {
		emitLet(bindings(expr), body(expr), si, env, isTail)
	} else if isLetrec(expr) {
		//emitLetrec(bindings(expr), body(expr), si, env)
		emitLetrec(expr)
	} else if isBegin(expr) {
		emitBegin(expr, si, env, isTail)
	} else if isPrimcall(expr) {
		emitPrimitiveCall(expr, si, env)
		emitRetIf(isTail)
	} else if isApp(expr, env) {
		emitApp(expr, si, env, isTail)
	} else {
		pp.Println(expr)
		panic(fmt.Sprintf("[emitAnyExpr] not implemented. %v", expr))
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
	emit("\t.globl %s", label)
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

	initPrimCallOpMap()

	compileProgram(target)
}
