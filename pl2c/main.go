package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/k0kubun/pp"
)

const (
	C_INT = iota
	C_STRING
	C_ARRAY
	C_FPTR
	C_PROC
	C_VAR
)

type ret struct {
	typeId int
	name   string
	proc   func([]*cell, *environment)
}

func newRet(typeId int, name string) ret {
	return ret{typeId, name, nil}
}

func newRetProc(procBody func([]*cell, *environment)) ret {
	// need environment?
	return ret{C_PROC, "proc", procBody}
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
	LB  = `\0A`
	EOF = `\00`
)

type environment struct {
	parent *environment
	dict   map[string]func([]*cell, *environment)

	count int

	strCount int
	header   map[string]int
	body     string

	retStack []ret
	defined  map[string]struct{}
}

func newLocalEnv(parent *environment) *environment {
	env := &environment{}
	env.parent = parent
	env.count = 0
	env.strCount = 0
	env.header = make(map[string]struct{})
	env.body = ""
	env.retStack = []ret{}
	env.defined = make(map[string]struct{})
	env.dict = map[string]func([]*cell, *environment){}
	return env
}

func newGlobalEnv() *environment {
	env := newLocalEnv(nil)

	env.dict["+"] = func(args []*cell, e *environment) {
		n := env.next()
		retName := fmt.Sprintf("plus_%d", n)
		env.addBody(fmt.Sprintf("int %s = 0;", retName))

		for _, c := range args {
			if c.isNum() {
				env.addBody(fmt.Sprintf("%s += %s;", retName, c.value))
			} else {
				emit(c, env)
				ret := env.popRet()
				env.addBody(fmt.Sprintf("%s += %s;", retName, ret.name))
			}
		}

		env.pushRet(newRet(C_INT, retName))
	}
	env.dict["atom"] = func(args []*cell, e *environment) {
		if len(args) != 1 {
			panic("atom: invalid arguments")
		}
		c := args[0]

		n := env.next()
		retName := fmt.Sprintf("atom_%d", n)
		if c.isAtom() || c.isNum() || c.isNil() {
			env.addBody(fmt.Sprintf("int %s = 1;", retName))
		} else {
			emit(c, env)

			ret := env.popRet()
			if ret.isArray() {
				// FIXME
				// At the compile time, it is impossible to determine whether
				// the returned value is atom or not.
				//env.addBody(fmt.Sprintf("int %s = 0 == ;", retName))
				env.addBody(fmt.Sprintf("int %s = 0;", retName))
			} else {
				env.addBody(fmt.Sprintf("int %s = 1;", retName))
			}
		}
		env.pushRet(newRet(C_INT, retName))
	}
	/*
		env.dict["lambda"] = func(args []*cell, e *environment) {
			n := env.next()
			retName := fmt.Sprintf("lambda_%d", n)
			env.pushRet(newRet(C_FPTR, retName))
		}
	*/
	env.dict["define"] = func(args []*cell, e *environment) {
		if args[0].isAtom() {
			// variable
			// simple assignment, When a variable is already declared.

			if !args[1].isNum() {
				panic("expected number as value")
			}
			retName := args[0].value
			value := args[1].value

			_, contains := env.defined[retName]
			if contains {
				env.addBody(fmt.Sprintf("%s = %s;", retName, value))
			} else {
				env.defined[retName] = struct{}{}
				// expect INT
				env.addBody(fmt.Sprintf("int %s = %s;", retName, value))
			}

			// maybe need care about a collision with system variables,
			// especially in case of invalid type.
			env.pushRet(newRet(C_STRING, retName))
		} else {
			// function
			panic("")
		}
	}
	env.dict["progn"] = func(args []*cell, e *environment) {
		// return integer now.

		var r ret
		for _, c := range args {
			emit(c, env)
			r = env.popRet()
		}

		n := env.next()
		retName := fmt.Sprintf("progn_%d", n)

		if r.isInt() {
			env.addBody(fmt.Sprintf("int %s = %s;", retName, r.name))
		} else {
			// FIXME
			env.addBody(fmt.Sprintf("int %s = 0;", retName))
		}

		env.pushRet(newRet(C_INT, retName))
	}
	env.dict["print"] = func(args []*cell, e *environment) {
		e.addHeader("stdio.h")

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
				emit(c, env)

				ret := env.popRet()
				if ret.isInt() {
					printArgs += fmt.Sprintf(", %s", ret.name)
					printBody += " %d"
				} else if ret.isString() {
					printArgs += fmt.Sprintf(", %s", ret.name)
					printBody += " %s"
				} else {
					panic("")
				}
			}
		}

		env.addBody(fmt.Sprintf("printf(\"%s\\n\" %s);", printBody[1:], printArgs))

		n := env.next()
		retName := fmt.Sprintf("print_ret_%d", n)
		env.addBody(fmt.Sprintf("char* %s = \"#<undef>\";", retName))

		env.pushRet(newRet(C_STRING, retName))
	}
	return env
}

func (this *environment) pushRet(r ret) {
	this.retStack = append(this.retStack, r)
}

func (this *environment) popRet() ret {
	n := len(this.retStack)
	ret := this.retStack[n-1]
	this.retStack = this.retStack[:n-1]
	return ret
}

func (this *environment) next() int {
	this.count++
	return this.count
}

func (this *environment) nextStr() int {
	this.strCount++
	return this.strCount
}

func (this *environment) addHeader(ir string) {
	if this.parent != nil {
		this.addHeaderGlobal(ir)
	} else {
		this.header[ir] = struct{}{}
	}
}

func (this *environment) addFuncBody(ir string) {
	this.funcBody += ir
	this.funcBody += "\n"
}

func (this *environment) addBody(ir string) {
	this.body += ir
	this.body += "\n"
}

//env.dict = map[string]func([]*cell, *environment){}
func (this *environment) find(name string) func([]*cell, *environment) {
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

func (this *environment) print() {
	for h, _ := range this.header {
		fmt.Printf("#include<%s>\n", h)
	}

	fmt.Println(`
typedef struct PAIR_ {
	int car;
	struct pair_ *cdr;
} PAIR;
	`)
	fmt.Println("int main() {")
	fmt.Println(this.body)
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
		exp := env.find(cell.value)
		if exp != nil {
			env.pushRet(newRetProc(exp))
			return
		}
		v := env.findVar(cell.value)
		if len(v) > 0 {
			env.pushRet(newRet(C_VAR, v))
			return
		}
		panic("not found: " + cell.value)

		//env.pushRet(newRet(C_INT, cell.value))
		return
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
		case "lambda":
			localEnv := newLocalEnv(env)
			for _, c := range cell.list[1].list {
				localEnv.defined[c.value] = struct{}{}
			}

			emit(cell.list[2], localEnv)
			pp.Println(localEnv)
			n := env.next()
			retName := fmt.Sprintf("lambda_%d", n)
			env.pushRet(newRet(C_FPTR, retName))
			return
		}
		//case "progn":
	default:
	}

	emit(cell.list[0], env)
	proc := env.popRet()
	exps := []ret{}
	for _, c := range cell.list[1:] {
		emit(c, env)
		exps = append(exps, env.popRet())
	}

	if proc.typeId == C_FPTR {
		labelArgs := ""
		for _, c := range exps {
			labelArgs += "," + c.name
		}
		env.addBody(proc.name + "(" + labelArgs[1:] + ");")
	} else if proc.typeId == C_PROC {
		pp.Println(env.dict)
		pp.Println(env.defined)
		pp.Println(env.retStack)
		panic("proc")
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
		list:   nil,
	}
}

func newNum(token string, n int) *cell {
	return &cell{
		typeId: LISP_NUM,
		value:  token,
		num:    n,
		list:   nil,
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
		list:   nil,
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

	/*
		code := `
		(progn

		(define a 10)
		(print (+ 3 a))
		)
		`
	*/
	//code := "(progn\n (define a 10) (print (+ 100 a)))"
	code := "((lambda (x) (+ x 1)) 10)"

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
