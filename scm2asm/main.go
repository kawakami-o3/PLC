package main

import (
	"flag"
	"fmt"
)

func emit(s string, v ...interface{}) {
	fmt.Println(fmt.Sprintf(s, v...))
}

func compileProgram(x string) {
	emit(".text")
	emit(".p2align 4,,15")
	emit(".globl scheme_entry")
	emit(".type scheme_entry, @function")
	emit("scheme_entry:")
	emit("movl %v %%eax", x)
	emit("ret")
}

func main() {
	flag.Parse()
	target := flag.Args()[0]

	compileProgram(target)
}
