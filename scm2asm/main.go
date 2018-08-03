package main

import (
	"flag"
	"fmt"
)

func emit(s string, v ...interface{}) {
	fmt.Println(fmt.Sprintf(s, v...))
}

func compileProgram(x string) {
	emit("\t.text")
	emit("\t.p2align 4,,15")
	emit("\t.globl scheme_entry")
	emit("\t.type scheme_entry, @function")
	emit("scheme_entry:")
	emit("\tmovl $%v, %%eax", x)
	emit("\tret")
}

func main() {
	flag.Parse()
	target := flag.Args()[0]

	compileProgram(target)
}
