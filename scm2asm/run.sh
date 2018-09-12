#!/bin/sh
set -eux

run() {
	./scm2asm.exe "$*"
}

goimports -w -l .
go build -o scm2asm.exe

#./scm2asm.exe '(null? ())' > sandbox/scheme_entry.s
#./scm2asm.exe '(char=? #\a #\c)'
#./scm2asm.exe '(let ((a 1)) a)'
#./scm2asm.exe '(letrec () 12)'
#./scm2asm.exe '(letrec () (let ([x 5]) (+ x x)))'
#run '(letrec ([f (lambda () 5)]) 7)'
run '(letrec ([f (lambda () 5)]) (f))'
