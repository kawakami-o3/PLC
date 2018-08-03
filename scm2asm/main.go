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
		return 1<<boolShift + boolTag
	case "#f":
		return 0<<boolShift + boolTag
	default:
		return int(x[0])<<charShift + charTag
	}
}

func compileProgram(x string) {
	emit("\t.text")
	emit("\t.p2align 4,,15")
	emit("\t.globl scheme_entry")
	emit("\t.type scheme_entry, @function")
	emit("scheme_entry:")
	emit("\tmovl $%d, %%eax", immediateRep(x))
	emit("\tret")
}

func main() {
	flag.Parse()
	target := flag.Args()[0]

	compileProgram(target)
}
