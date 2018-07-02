package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	C_INT = iota
	C_STRING
	C_ARRAY
	C_FPTR
	C_PROC
	C_VAR
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

func (this ret) isInt() bool {
	return this.typeId == C_INT
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
	/*
		dictFunc   map[string]func([]*cell, *environment)
		dictVar  map[string]struct{}
	*/
	dict map[string]*decl

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

	//env.dictFunc = map[string]func([]*cell, *environment){}
	//env.dictVar = make(map[string]struct{})
	env.dict = map[string]*decl{}

	env.count = 0
	env.strCount = 0

	env.header = make(map[string]struct{})
	env.pre = ""
	env.main = ""

	env.retStack = []ret{}
	return env
}

func newGlobalEnv() *environment {
	env := newLocalEnv(nil)

	env.registFunc("+", func(args []*cell, local *environment) {
		n := local.next()
		retName := fmt.Sprintf("plus_%d", n)
		local.putsMain(fmt.Sprintf("int %s = 0;", retName))

		//for _, c := range args {
		for i := 0; i < len(args); i++ {
			ret := local.popRet()
			local.putsMain(fmt.Sprintf("%s += %s;", retName, ret.name))
			/*
				if c.isNum() {
					local.putsMain(fmt.Sprintf("%s += %s;", retName, c.value))
				} else {
					//emit(c, local)
					ret := local.popRet()
					local.putsMain(fmt.Sprintf("%s += %s;", retName, ret.name))
				}
			*/
		}

		local.pushRet(newRet(C_INT, retName))
	})
	env.registFunc("-", func(args []*cell, local *environment) {
		n := local.next()
		retName := fmt.Sprintf("minus_%d", n)
		ret := local.popRet()
		local.putsMain(fmt.Sprintf("int %s = %s;", retName, ret.name))

		//for _, c := range args {
		for i := 1; i < len(args); i++ {
			ret := local.popRet()
			local.putsMain(fmt.Sprintf("%s -= %s;", retName, ret.name))
		}

		local.pushRet(newRet(C_INT, retName))
	})
	env.registFunc("atom", func(args []*cell, e *environment) {
		if len(args) != 1 {
			panic("atom: invalid arguments")
		}
		c := args[0]

		n := env.next()
		retName := fmt.Sprintf("atom_%d", n)
		if c.isAtom() || c.isNum() || c.isNil() {
			env.putsMain(fmt.Sprintf("int %s = 1;", retName))
		} else {
			emit(c, env)

			ret := env.popRet()
			if ret.isArray() {
				// FIXME
				// At the compile time, it is impossible to determine whether
				// the returned value is atom or not.
				//env.putsMain(fmt.Sprintf("int %s = 0 == ;", retName))
				env.putsMain(fmt.Sprintf("int %s = 0;", retName))
			} else {
				env.putsMain(fmt.Sprintf("int %s = 1;", retName))
			}
		}
		env.pushRet(newRet(C_INT, retName))
	})
	env.registFunc("eq", func(args []*cell, e *environment) {
		//e.include("stdbool.h")

		rets := []*ret{}
		for i := 0; i < len(args); i++ {
			//emit(c, e)
			rets = append(rets, e.popRet())
		}

		first := rets[0].name
		exps := []string{}
		for _, r := range rets[1:] {
			exps = append(exps, fmt.Sprintf("%s == %s", first, r.name))
		}

		n := e.next()
		retName := fmt.Sprintf("eq_%d", n)

		//e.putsMain(fmt.Sprintf("bool %s = (", retName))
		e.putsMain(fmt.Sprintf("int %s = (", retName))
		e.putsMain(strings.Join(exps, " && "))
		e.putsMain(");")
		e.pushRet(newRet(C_INT, retName))
	})
	env.registFunc("print", func(args []*cell, e *environment) {
		e.include("stdio.h")

		printArgs := ""
		printBody := ""
		for _, c := range args {
			if c.isNum() {
				printArgs += fmt.Sprintf(", %s", c.value)
				printBody += " %d"
			} else if c.isAtom() {
				printArgs += fmt.Sprintf(", %s", c.value)
				printBody += " %d"
			} else {
				/*
					emit(c, env)
				*/

				ret := e.popRet()
				if ret.isInt() {
					printArgs += fmt.Sprintf(", %s", ret.name)
					printBody += " %d"
				} else if ret.isString() {
					printArgs += fmt.Sprintf(", %s", ret.name)
					printBody += " %s"
				} else if ret.isArray() {

					printBody += " ("
					for i := 0; i < ret.length; i++ {
						printArgs += fmt.Sprintf(", %s[%d]", ret.name, i)
						printBody += " %d"
					}
					printBody += " )"
				} else {
					panic("")
				}
			}
		}

		e.putsMain(fmt.Sprintf("printf(\"%s\\n\" %s);", printBody[1:], printArgs))

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

func (this *environment) pushRet(r ret) {
	this.retStack = append(this.retStack, r)
}

func (this *environment) popRet() *ret {
	n := len(this.retStack)
	/*
		if n == 0 {
			return nil
		}
	*/
	ret := this.retStack[n-1]
	this.retStack = this.retStack[:n-1]
	return &ret
}

func (this *environment) next() int {
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
	/*
		if this.parent == nil {
			this.main += ir
			this.main += "\n"
		} else {
			this.parent.putsMain(ir)
		}
	*/
}

//env.dict = map[string]func([]*cell, *environment){}
//func (this *environment) find(name string) func([]*cell, *environment) {
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

/*
func (this *environment) findVar(name string) string {
	_, contains := this.defined[name]
	if contains {
		return name
	} else {
		if this.parent == nil {
			panic("not found: " + name)
		} else {
			return this.parent.findVar(name)
		}
	}
}
*/

func (this *environment) print() {
	for h, _ := range this.header {
		fmt.Printf("#include<%s>\n", h)
	}
	/*
	   	fmt.Println(`
	   typedef struct PAIR_ {
	   	int car;
	   	struct pair_ *cdr;
	   } PAIR;
	   	`)
	*/
	fmt.Println(this.pre)
	fmt.Println("int main() {")
	fmt.Println(this.main)
	fmt.Println("}")
}

func emit(cell *cell, env *environment) {
	switch cell.typeId {
	case LISP_NUM:
		env.pushRet(newRet(C_INT, cell.value))
		return
	case LISP_ATOM:
		// value or function
		//env.pushRet(newRet(C_FPTR, env.find(cell.value)))
		d := env.find(cell.value)
		if d != nil {
			switch d.typeId {
			case DECL_FUNC:
				env.pushRet(newRetProc(d.name, d.proc, env))
			case DECL_VAR:
				env.pushRet(newRet(C_VAR, d.name))
			}
			return
		}
		//panic("not found: " + cell.value)

		//env.pushRet(newRet(C_INT, cell.value))
		//return
	case LISP_LIST:
		head := cell.list[0].value
		/*
			if head.typeId == LISP_ATOM {
				emitter := env.dict[cell.list[0].value]
				if emitter == nil {
					fmt.Println("-", cell.list[0].value, "-")
					panic("")
				}

				emitter(cell.list[1:], env)
				return
			} else {
		*/
		switch head {
		case "quote":
			// int and flat int array are implemented. The other raises panic.
			n := env.next()
			retName := fmt.Sprintf("quote_%d", n)

			if len(cell.list) > 2 {
				panic("quote: too many arguments")
			}

			if cell.list[1].typeId == LISP_LIST {
				arrBody := ""
				for _, i := range cell.list[1].list {
					arrBody += "," + i.value
				}
				env.putsMain(fmt.Sprintf("int %s[] = { %s };", retName, arrBody[1:]))

				ret := newRet(C_ARRAY, retName)
				ret.length = len(cell.list[1].list)

				env.pushRet(ret)
			} else {
				env.putsMain(fmt.Sprintf("int %s = %s;", retName, cell.list[1].value))
				env.pushRet(newRet(C_INT, retName))
			}
			return
		case "if":
			n := env.next()
			retName := fmt.Sprintf("if_%d", n)
			env.putsMain(fmt.Sprintf("int %s;", retName))

			emit(cell.list[1], env)
			cnd := env.popRet()

			env.putsMain(fmt.Sprintf("if (%s) {", cnd.name))
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
			if body.typeId == LISP_LIST && len(body.list) > 0 && body.list[0].value == "lambda" {

				retName := label.value
				d := env.find(retName)
				if d == nil {
					env.registFunc(retName, nil) // need body?
				}

				env.putsPre(fmt.Sprintf("int (*%s)(int*);", retName))
				emit(body, env)
				bodyRet := env.popRet()
				env.putsMain(fmt.Sprintf("%s = %s;", retName, bodyRet.name))

				/*
					if d == nil {
						// expect INT
						//env.putsMain(fmt.Sprintf("int (*%s)(int*) = %s;", retName, bodyRet.name))
						env.putsPre(fmt.Sprintf("int (*%s)(int*) = %s;", retName, bodyRet.name))
					} else {
						env.putsMain(fmt.Sprintf("%s = %s;", retName, bodyRet.name))
					}
				*/
				env.pushRet(newRet(C_STRING, retName))

			} else if body.isNum() {
				retName := label.value
				value := body.value

				d := env.find(retName)
				if d == nil {
					env.registVar(retName)
					// expect INT
					env.putsMain(fmt.Sprintf("int %s = %s;", retName, value))
				} else {
					env.putsMain(fmt.Sprintf("%s = %s;", retName, value))
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

			env.putsPre(fmt.Sprintf("int %s(int *args){", retName))
			for i, c := range cell.list[1].list {
				env.putsPre(fmt.Sprintf("int %s = args[%d];", c.value, i))
			}
			env.putsPre(localEnv.main)
			env.putsPre(fmt.Sprintf("return %s;", localRet.name))
			env.putsPre("}")

			env.pushRet(ret)
			return
		}
		//case "progn":
	default:
	}

	emit(cell.list[0], env)
	proc := env.popRet()
	exps := []*ret{}
	for _, c := range cell.list[1:] {
		emit(c, env)
		exps = append(exps, env.popRet())
	}

	if proc.typeId == C_FPTR {
		labelArgs := ""
		for _, c := range exps {
			labelArgs += "," + c.name
		}
		env.putsMain(proc.name + "(" + labelArgs[1:] + ");")
	} else if proc.typeId == C_PROC {
		//if proc.name == "proc" {
		if proc.proc != nil {
			// proc
			for _, e := range exps {
				env.retStack = append([]ret{*e}, env.retStack...)
			}
			proc.proc(cell.list[1:], env)
		} else {
			// lambda
			/*
				ret := proc.env.popRet()

				env.putsPre(fmt.Sprintf("int %s(int *args){", proc.name))
				for i, name := range proc.args {
					env.putsPre(fmt.Sprintf("int %s = args[%d];", name, i))
				}
				env.putsPre(proc.env.main)
				env.putsPre(fmt.Sprintf("return %s;", ret.name))
				env.putsPre("}")
			*/

			n := env.next()
			argName := fmt.Sprintf("%s_args_%d", proc.name, n)
			env.putsMain(fmt.Sprintf("int %s[] = {", argName)) // declaration
			argCnt := ""
			for _, e := range exps {
				argCnt += fmt.Sprintf(",%s", e.name)
			}
			env.putsMain(argCnt[1:])
			env.putsMain("};")
			env.pushRet(newRet(C_INT, fmt.Sprintf("%s(%s)", proc.name, argName)))

		}

	} else {
		panic("not a function: " + proc.name)
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

//func newAtom(token string) (*cell, error) {
func newAtom(token string) *cell {
	/*
		v, err := strconv.Atoi(token)
		if err != nil {
			return nil, err
		}
	*/
	return &cell{
		typeId: LISP_ATOM,
		value:  token,
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
		if c == '\n' {
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

/*
func parse(code string) *cell {
	cell, err := readFrom(tokenize(code))
	if err != nil {
		panic(err)
	}
	return cell
}
*/

func main() {
	//code := "(+ 1 (- 3 2))\n(+ 1 (- 3 2))"
	//code := "(+ 1 (- 3 2))"
	//code := "(+ 1 2)"
	//code := "(print (print 1 2 3 4) (print 3 5 5))"
	//code := "(print (atom 4) (atom (print 3)))"
	//code := "(print (+ 1 2))"
	//code := "(print (+ 3 (+ 1 2)))"
	//code := "(print (+ 3 (+ 1 2) 8))" // FIXME raise compile error
	//code := "(print (+ 3 (+ 1 2) 8)))"

	//code := "(progn\n (define a 10) (print (+ 100 a)))"

	// lambda
	//code := "(print ((lambda (x) (+ x 1)) 10))"
	//code := "(print ((lambda (x) (+ x 1)) 10) ((lambda (y) (+ y 20)) 3))"
	//code := "(print ((lambda (x) (+ x 1)) 10) ((lambda (y z) (+ y z 20)) 3 8))"

	// define lambda
	//code := "(define fact 10)" // define function -> special form
	//code := "(progn (define fact (lambda (x) (+ x 1))) (print (fact 10)))"

	// eq
	//code := "(print (eq 1 1))"
	//code := "(print (eq 1 0))"
	//code := "(print (eq 1 1 (+ 1 3)))"

	// if
	//code := "(print (if (eq 1 1) 1 0))"
	//code := "(progn (define fib (lambda (n) (if (eq n 0) 1 (if (eq n 1) 1 (+ (fib (- n 1)) (fib (- n 2))))))) (print (fib 1)))"
	//code := "(progn (define fib (lambda (n) (if (eq n 0) 1 (if (eq n 1) 1 (+ (fib (- n 1)) (fib (- n 2))))))) (print (fib 2)) (print (fib 3)) (print (fib 4)) (print (fib 5)) (print (fib 6)))"

	// quote
	//code := "(print (quote 1))"
	code := "(print (quote (1 2 3)))"

	/*
		fmt.Println(code)
		fmt.Println(parse(code).str())
		fmt.Println(compile(code))
	*/

	cell, err := readFrom(tokenize(code))
	if err != nil {
		panic(err)
	}

	env := newGlobalEnv()
	emit(cell, env)

	env.print()

}
