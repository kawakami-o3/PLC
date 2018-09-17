#!/bin/sh
set -eux

run() {
	./scm2asm.exe -- "$*"
}

goimports -w -l .
go build -o scm2asm.exe

#./scm2asm.exe '(null? ())' > sandbox/scheme_entry.s
#./scm2asm.exe '(char=? #\a #\c)'
#./scm2asm.exe '(let ((a 1)) a)'
#./scm2asm.exe '(letrec () 12)'
#./scm2asm.exe '(letrec () (let ([x 5]) (+ x x)))'
#run '(letrec ([f (lambda () 5)]) 7)'
#run '(letrec ([f (lambda () 5)]) (f))'
#run '(letrec ([f (lambda () (+ 5 7))] [g (lambda () 13)]) (+ (f) (g)))'

run '(letrec ([f (lambda (x) (+ x 1))]) (f 2))'
#run '(letrec ([f (lambda (x) (+ x 12))] [g (lambda (x) (* x (g 2)))]) (g (f 13)))'
#run '(letrec ([f (lambda (x) (+ x 1))] [g (lambda (x) (* (f x) 2))]) (g (g (g 3))))'

#run '(letrec ([f (lambda (x) (if (zero? x) 0 (+ 1 (f (sub1 x)))))]) (f 200))'
#run '(letrec ([sum (lambda (n ac) (if (zero? n) ac (app sum (sub1 n) (+ n ac))))]) (sum 10 0))'

#run '(letrec ([e (lambda (x) (if (zero? x) #t (o (sub1 x))))]
#            [o (lambda (x) (if (zero? x) #f (e (sub1 x))))])
#     (e 25))'
 

