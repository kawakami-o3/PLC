#!/bin/sh
set -eux

goimports -w -l .
go build -o scm2asm.exe

#./scm2asm.exe '(null? ())' > sandbox/scheme_entry.s
#./scm2asm.exe '(char=? #\a #\c)'
#./scm2asm.exe '(let ((a 1)) a)'
./scm2asm.exe '(letrec () 12)'


