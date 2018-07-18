package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func readFile(filename string) string {
	//file, err := os.Open("templates/common.c")
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}

const (
	C_INT = iota
	C_STRING
	C_ARRAY
	C_FPTR
	C_PROC
	C_VAR
	C_UNKNOWN
	C_CALL
)

const (
	LB  = `\0A`
	EOF = `\00`
)

type ret struct {
	typeId int
	name   string
	proc   func([]*cell, *environment)
	env    *environment
	args   []string

	length int // for 'quote' and 'print'
}

func newRet(typeId int, name string) ret {
	return ret{typeId, name, nil, nil, []string{}, 0}
}

func newRetProc(name string, proc func([]*cell, *environment), env *environment) ret {
	return ret{C_PROC, name, proc, env, []string{}, 0}
}

func newRetArr(name string, length int) ret {
	return ret{C_ARRAY, name, nil, nil, []string{}, length}
}

func (this ret) isInt() bool {
	return this.typeId == C_INT
}

func (this ret) isCall() bool {
	return this.typeId == C_CALL
}

func (this ret) isString() bool {
	return this.typeId == C_STRING
}

func (this ret) isArray() bool {
	return this.typeId == C_ARRAY
}

const (
	DECL_FUNC = iota
	DECL_VAR
)

type decl struct {
	typeId int
	name   string                      // variable name
	proc   func([]*cell, *environment) // proc body
}

func newDecl(typeId int, name string, proc func([]*cell, *environment)) *decl {
	return &decl{typeId, name, proc}
}

type environment struct {
	parent *environment

	dict     map[string]*decl
	external []string

	// not yet shared in an environment tree.
	count    int // instruction count
	strCount int

	header map[string]struct{} // header file
	pre    string
	main   string

	retStack []ret
}

func newLocalEnv(parent *environment) *environment {
	env := &environment{}
	env.parent = parent

	env.dict = map[string]*decl{}
	env.external = []string{}

	env.count = 0
	env.strCount = 0

	env.header = make(map[string]struct{})
	env.pre = ""
	env.main = ""

	env.retStack = []ret{}
	return env
}

func consAll(names []string) string {
	ret := ""
	for _, n := range names {
		ret += fmt.Sprintf("cons(%s, ", n)
	}
	ret += "nil"
	for i := 0; i < len(names); i++ {
		ret += ")"
	}
	return ret
}

func newGlobalEnv() *environment {
	env := newLocalEnv(nil)

	env.registVar("nil")
	env.registVar("t")

	env.registFunc("+", func(args []*cell, local *environment) {
		n := local.next()
		retName := fmt.Sprintf("plus_%d", n)
		argsName := fmt.Sprintf("args_%d", n)

		argNames := []string{}
		for i := 0; i < len(args); i++ {
			ret := local.popRet()
			argNames = append(argNames, ret.name)
		}
		local.putsMain(fmt.Sprintf("List *%s = %s;", argsName, consAll(argNames)))

		local.putsMain(fmt.Sprintf("List *%s = plc_add(%s);", retName, argsName))

		local.pushRet(newRet(C_INT, retName))
	})
	env.registFunc("-", func(args []*cell, local *environment) {
		n := local.next()
		retName := fmt.Sprintf("minus_%d", n)
		argsName := fmt.Sprintf("args_%d", n)

		argNames := []string{}
		for i := 0; i < len(args); i++ {
			ret := local.popRet()
			argNames = append(argNames, ret.name)
		}
		local.putsMain(fmt.Sprintf("List *%s = %s;", argsName, consAll(argNames)))

		local.putsMain(fmt.Sprintf("List *%s = plc_sub(%s);", retName, argsName))

		local.pushRet(newRet(C_INT, retName))
	})
	env.registFunc("atom", func(args []*cell, e *environment) {
		if len(args) != 1 {
			panic("atom: invalid arguments")
		}
		n := e.next()
		retName := fmt.Sprintf("atom_%d", n)
		arg := e.popRet()

		e.putsMain(fmt.Sprintf("List *%s = make_int(%s->atom != NULL);", retName, arg.name))
		e.pushRet(newRet(C_INT, retName))
	})
	env.registFunc("eq", func(args []*cell, e *environment) {
		rets := []*ret{}
		for i := 0; i < len(args); i++ {
			rets = append(rets, e.popRet())
		}

		exps := []string{}
		for _, r := range rets {
			exps = append(exps, fmt.Sprintf("%s", r.name))
		}

		n := e.next()
		retName := fmt.Sprintf("eq_%d", n)

		e.putsMain(fmt.Sprintf("List *%s = eq(", retName))
		e.putsMain(strings.Join(exps, ","))
		e.putsMain(");")
		e.pushRet(newRet(C_INT, retName))
	})
	env.registFunc("car", func(args []*cell, e *environment) {
		arg := e.popRet()

		n := e.next()
		retName := fmt.Sprintf("car_%d", n)
		e.putsMain(fmt.Sprintf("List *%s = car(%s);", retName, arg.name))
		e.pushRet(newRet(C_INT, retName))
	})
	env.registFunc("cdr", func(args []*cell, e *environment) {
		arg := e.popRet()

		n := e.next()
		retName := fmt.Sprintf("cdr_%d", n)

		e.putsMain(fmt.Sprintf("List *%s = cdr(%s);", retName, arg.name))
		e.pushRet(newRetArr(retName, arg.length-1))
	})
	env.registFunc("cons", func(args []*cell, e *environment) {
		elm := e.popRet()
		rest := e.popRet()

		n := e.next()
		retName := fmt.Sprintf("cons_%d", n)
		e.putsMain(fmt.Sprintf("List *%s = cons(%s, %s);", retName, elm.name, rest.name))
		e.pushRet(newRetArr(retName, 1+rest.length))
	})
	env.registFunc("print", func(args []*cell, e *environment) {
		for i := 0; i < len(args); i++ {
			ret := e.popRet()
			e.putsMain(fmt.Sprintf("printList(%s);", ret.name))
		}

		n := e.next()
		retName := fmt.Sprintf("print_ret_%d", n)
		e.putsMain(fmt.Sprintf("char* %s = \"#<undef>\";", retName))

		e.pushRet(newRet(C_STRING, retName))
	})
	return env
}

func (this *environment) registFunc(name string, proc func(args []*cell, e *environment)) {
	this.dict[name] = newDecl(DECL_FUNC, name, proc)
}

func (this *environment) registVar(name string) {
	this.dict[name] = newDecl(DECL_VAR, name, nil)
}

func (this *environment) registGlobal(name string) {
	if this.parent != nil {
		this.parent.registGlobal(name)
	}
	this.dict[name] = newDecl(DECL_VAR, name, nil)
}

func (this *environment) pushRet(r ret) {
	this.retStack = append(this.retStack, r)
}

func (this *environment) popRet() *ret {
	n := len(this.retStack)
	ret := this.retStack[n-1]
	this.retStack = this.retStack[:n-1]
	return &ret
}

func (this *environment) next() int {
	if this.parent != nil {
		return this.parent.next()
	}
	this.count++
	return this.count
}

func (this *environment) nextStr() int {
	this.strCount++
	return this.strCount
}

func (this *environment) include(ir string) {
	if this.parent == nil {
		this.header[ir] = struct{}{}
	} else {
		this.include(ir)
	}
}

func (this *environment) putsPre(ir string) {
	if this.parent == nil {
		this.pre += ir
		this.pre += "\n"
	} else {
		this.parent.putsPre(ir)
	}
}

func (this *environment) putsMain(ir string) {
	this.main += ir
	this.main += "\n"
}

func (this *environment) find(name string) *decl {
	ret, contains := this.dict[name]
	if contains {
		return ret
	} else {
		if this.parent == nil {
			return nil
		} else {
			return this.parent.find(name)
		}
	}
}

func (this *environment) have(name string) bool {
	_, contains := this.dict[name]
	return contains
}

func (this *environment) hasGlobal(name string) bool {
	if this.parent != nil {
		return this.parent.hasGlobal(name)
	}

	elm := this.dict[name]
	if elm == nil || elm.typeId == DECL_FUNC {
		return false
	} else {
		return true
	}
}

func (this *environment) globals() []string {
	if this.parent != nil {
		return this.parent.globals()
	}
	names := []string{}
	for k, _ := range this.dict {
		names = append(names, k)
	}
	return names
}

func (this *environment) registExternal(name string) {
	if name == "nil" || name == "t" {
		return
	}
	for _, s := range this.external {
		if s == name {
			return
		}
	}
	for _, s := range this.globals() {
		if s == name {
			return
		}
	}
	this.external = append(this.external, name)
}

func (this *environment) print() {
	for h, _ := range this.header {
		fmt.Printf("#include<%s>\n", h)
	}

	fmt.Println(readFile("templates/common.c"))

	fmt.Println(this.pre)
	fmt.Println("int main() {")
	fmt.Println("init_common();")
	fmt.Println(this.main)
	fmt.Println("}")
}

var sanitizeMatchers map[string]*regexp.Regexp

func sanitizeName(name string) string {
	if len(name) == 1 {
		return name
	}
	if sanitizeMatchers == nil {
		sanitizeMatchers = map[string]*regexp.Regexp{}

		sanitizeMatchers["_s_"] = regexp.MustCompile(`\*`)
		sanitizeMatchers["_k_"] = regexp.MustCompile(`:`) // keyword
		sanitizeMatchers["_m_"] = regexp.MustCompile(`-`) // FIXME (- 1 1)
	}

	for k, m := range sanitizeMatchers {
		name = m.ReplaceAllString(name, k)
	}
	return name
}

func emitSymbol(cell *cell, env *environment) {
	n := env.next()
	symName := fmt.Sprintf("sym_%d", n)
	if env.hasGlobal(cell.value) {
		env.putsMain(fmt.Sprintf("List *%s = %s;", symName, cell.value))
	} else {
		strName := fmt.Sprintf("str_%d", n)

		env.putsMain(fmt.Sprintf("char *%s = \"%s\";", strName, cell.value))
		env.putsMain(fmt.Sprintf("List *%s = make_symbol(%s);", symName, strName))
	}
	env.pushRet(newRet(C_STRING, symName))
}

func emit(cell *cell, env *environment) {
	switch cell.typeId {
	case LISP_NUM:
		n := env.next()
		retName := fmt.Sprintf("num_%d", n)
		env.putsMain(fmt.Sprintf("List *%s = make_int(%s);", retName, cell.value))
		env.pushRet(newRet(C_INT, retName))
		return
	case LISP_ATOM:
		// value or function
		d := env.find(cell.value)
		if d != nil {
			switch d.typeId {
			case DECL_FUNC:
				env.pushRet(newRetProc(d.name, d.proc, env))
				return
			case DECL_VAR:
				if !env.have(cell.value) {
					env.registExternal(cell.value)
				}
				env.pushRet(newRet(C_VAR, d.name))
				return
			}
		}

		//env.pushRet(newRet(C_UNKNOWN, cell.value))
		panic(fmt.Sprintf("unknown!!! %s", cell.value))
		return
	case LISP_LIST:
		head := cell.list[0].value
		switch head {
		case "quote":
			// int and flat int array are implemented. The other raises panic.
			n := env.next()
			retName := fmt.Sprintf("quote_%d", n)

			if len(cell.list) > 2 {
				panic("quote: too many arguments")
			}

			if cell.list[1].typeId == LISP_LIST {
				names := []string{}
				for _, i := range cell.list[1].list {
					if i.typeId == LISP_NUM {
						emit(i, env)
					} else {
						emitSymbol(i, env)
					}
					r := env.popRet()
					names = append(names, r.name)
				}

				env.putsMain(fmt.Sprintf("List *%s = %s;", retName, consAll(names)))

				ret := newRet(C_ARRAY, retName)
				ret.length = len(cell.list[1].list)

				env.pushRet(ret)
			} else if cell.list[1].typeId == LISP_NUM {
				env.putsMain(fmt.Sprintf("List *%s = make_int(%s);", retName, cell.list[1].value))
				env.pushRet(newRet(C_INT, retName))
			} else if cell.list[1].typeId == LISP_ATOM {
				env.putsMain(fmt.Sprintf("List *%s = make_symbol(\"%s\");", retName, cell.list[1].value))
				env.pushRet(newRet(C_STRING, retName))
			} else {
				panic("quote: not implemented to " + cell.list[1].value)
			}
			return
		case "if":
			n := env.next()
			retName := fmt.Sprintf("if_%d", n)
			env.putsMain(fmt.Sprintf("List *%s;", retName))

			emit(cell.list[1], env)
			cnd := env.popRet()

			env.putsMain(fmt.Sprintf("if (%s->atom->i) {", cnd.name))
			emit(cell.list[2], env)
			tRet := env.popRet()
			env.putsMain(fmt.Sprintf("%s = %s;", retName, tRet.name))
			env.putsMain("} else {")
			emit(cell.list[3], env)
			fRet := env.popRet()
			env.putsMain(fmt.Sprintf("%s = %s;", retName, fRet.name))
			env.putsMain("}")

			env.pushRet(newRet(C_INT, retName))

			return
		case "progn":
			var r *ret
			for _, c := range cell.list[1:] {
				emit(c, env)
				r = env.popRet()
				if r.isCall() {
					env.putsMain(fmt.Sprintf("%s;", r.name))
				}
			}

			n := env.next()
			retName := fmt.Sprintf("progn_%d", n)

			if r.isInt() {
				env.putsMain(fmt.Sprintf("int %s = %s;", retName, r.name))
			} else {
				// FIXME
				env.putsMain(fmt.Sprintf("int %s = 0;", retName))
			}

			env.pushRet(newRet(C_INT, retName))
			return
		case "define":
			//args := cell.list[1:]
			label := cell.list[1]
			body := cell.list[2]
			if body.typeId == LISP_LIST {
				if len(body.list) > 0 && body.list[0].value == "lambda" {

					retName := label.value
					d := env.find(retName)
					if d == nil {
						env.registFunc(retName, nil) // need body?
						env.putsPre(fmt.Sprintf("List *%s;", retName))
					}

					emit(body, env)
					bodyRet := env.popRet()

					env.putsMain(fmt.Sprintf("%s = %s;", retName, bodyRet.name))

					env.pushRet(newRet(C_STRING, retName))
				} else {
					emit(body, env)
					bodyRet := env.popRet()

					retName := label.value
					d := env.find(retName)
					if d == nil {
						env.registGlobal(retName)
						env.putsPre(fmt.Sprintf("List *%s;", retName))
					}
					env.putsMain(fmt.Sprintf("%s = %s;", retName, bodyRet.name))

					env.pushRet(newRet(C_STRING, retName))
				}
			} else if body.isNum() {
				retName := label.value
				value := body.value

				d := env.find(retName)
				if d == nil {
					env.registVar(retName)
					// expect INT
					env.putsMain(fmt.Sprintf("List *%s = make_int(%s);", retName, value))
				} else {
					env.putsMain(fmt.Sprintf("%s = make_int(%s);", retName, value))
				}

				// maybe need care about a collision with system variables,
				// especially in case of invalid type.
				env.pushRet(newRet(C_STRING, retName))

			} else {
				panic("not implemented [0]")
			}
			return
		case "lambda":
			localEnv := newLocalEnv(env)
			for _, c := range cell.list[1].list {
				localEnv.registVar(c.value)
			}

			emit(cell.list[2], localEnv)
			n := env.next()
			retName := fmt.Sprintf("lambda_%d", n)
			ret := newRetProc(retName, nil, localEnv)
			for _, c := range cell.list[1].list {
				ret.args = append(ret.args, c.value)
			}
			localRet := localEnv.popRet()

			procName := fmt.Sprintf("lambda_proc_%d", n)

			if len(localEnv.external) > 0 {
				env.putsPre(fmt.Sprintf("List *%s_external;", procName))
			}
			env.putsPre(fmt.Sprintf("List *%s(List *args){", procName))
			for i, c := range cell.list[1].list {
				env.putsPre(fmt.Sprintf("List *%s = nth(args, %d);", c.value, i))
			}
			for i, s := range localEnv.external {
				env.putsPre(fmt.Sprintf("List *%s = nth(%s_external, %d);", s, procName, i))
			}
			env.putsPre(localEnv.main)
			env.putsPre(fmt.Sprintf("return %s;", localRet.name))
			env.putsPre("}")

			if len(localEnv.external) > 0 {
				env.putsMain(fmt.Sprintf("%s_external = %s;", procName, consAll(localEnv.external)))
			}
			env.putsMain(fmt.Sprintf("List *%s = make_lambda(%s);", retName, procName))

			env.pushRet(ret)
			return
		}
	default:
	}

	emit(cell.list[0], env)
	proc := env.popRet()
	exps := []*ret{}
	for _, c := range cell.list[1:] {
		emit(c, env)
		r := env.popRet()
		exps = append(exps, r)
	}

	if proc.typeId == C_FPTR {
		labelArgs := ""
		for _, c := range exps {
			labelArgs += "," + c.name
		}
		env.putsMain(proc.name + "(" + labelArgs[1:] + ");")
	} else if proc.typeId == C_PROC {
		if proc.proc != nil {
			// proc
			for _, e := range exps {
				env.retStack = append([]ret{*e}, env.retStack...)
			}
			proc.proc(cell.list[1:], env)
		} else {
			n := env.next()
			argName := fmt.Sprintf("%s_args_%d", proc.name, n)
			expNames := []string{}
			for _, e := range exps {
				expNames = append(expNames, e.name)
			}

			env.putsMain(fmt.Sprintf("List *%s = %s;", argName, consAll(expNames))) // declaration
			env.pushRet(newRet(C_CALL, fmt.Sprintf("%s->atom->proc(%s)", proc.name, argName)))
		}

	} else {
		// FIXME
		//panic("not a function: " + proc.name)
		n := env.next()
		argName := fmt.Sprintf("%s_args_%d", proc.name, n)
		expNames := []string{}
		for _, e := range exps {
			expNames = append(expNames, e.name)
		}

		env.putsMain(fmt.Sprintf("List *%s = %s;", argName, consAll(expNames))) // declaration
		//env.putsMain("// hoge")

		env.pushRet(newRet(C_INT, fmt.Sprintf("%s->atom->proc(%s)", proc.name, argName)))
	}
}

const (
	LISP_ATOM = iota
	LISP_NUM
	LISP_LIST
	LISP_LAMBDA
)

type cell struct {
	typeId int
	value  string
	num    int
	list   []*cell

	proc func(*cell) *cell
}

func newAtom(token string) *cell {
	return &cell{
		typeId: LISP_ATOM,
		value:  sanitizeName(token),
		list:   []*cell{},
	}
}

func newNum(token string, n int) *cell {
	return &cell{
		typeId: LISP_NUM,
		value:  token,
		num:    n,
		list:   []*cell{},
	}
}

func newList() *cell {
	return &cell{
		typeId: LISP_LIST,
		value:  "",
		list:   []*cell{},
	}
}

func newLambda() *cell {
	return &cell{
		typeId: LISP_LAMBDA,
		value:  "",
		list:   []*cell{},
	}
}

func (this *cell) str() string {
	switch this.typeId {
	case LISP_ATOM:
		return this.value
	default:
		ret := "["
		for _, i := range this.list {
			ret += i.str() + ", "
		}
		ret += "]"
		return ret
	}
}

func (this *cell) isAtom() bool {
	return this.typeId == LISP_ATOM
}

func (this *cell) isNum() bool {
	return this.typeId == LISP_NUM
}

func (this *cell) isNil() bool {
	return this.typeId == LISP_LIST && len(this.list) == 0
}

type TokenBuffer struct {
	tokens []string
	idx    int
}

func (this *TokenBuffer) get() string {
	return this.tokens[this.idx]
}

func (this *TokenBuffer) next() string {
	this.idx++
	return this.get()
}

func tokenize(code string) *TokenBuffer {
	tokens := []string{}
	t := ""
	for _, c := range code {
		if c == '\n' || c == '\t' {
			c = ' '
		}
		if c == ' ' && len(t) > 0 {
			tokens = append(tokens, t)
			t = ""
		} else if c == '(' || c == ')' {
			if len(t) > 0 {
				tokens = append(tokens, t)
				t = ""
			}
			tokens = append(tokens, string(c))
		} else if c != ' ' {
			t += string(c)
		}
	}
	return &TokenBuffer{
		tokens: tokens,
		idx:    0,
	}
}

func readFrom(buf *TokenBuffer) (*cell, error) {
	t := buf.get()
	if t == "(" {
		ret := newList()
		for buf.next() != ")" {
			cell, err := readFrom(buf)
			if err != nil {
				return nil, err
			}
			ret.list = append(ret.list, cell)
		}
		return ret, nil
	} else if t == ")" {
		return nil, errors.New("unexpected ')'")
	} else {
		n, err := strconv.Atoi(t)
		if err == nil {
			return newNum(t, n), nil
		} else {
			return newAtom(t), nil
		}
	}
}

func main() {
	flag.Parse()
	srcPath, err := filepath.Abs(flag.Args()[0])
	if err != nil {
		panic(err)
	}
	srcCnt := readFile(srcPath)

	cell, err := readFrom(tokenize(srcCnt))
	if err != nil {
		panic(err)
	}

	env := newGlobalEnv()
	emit(cell, env)

	env.print()
}
