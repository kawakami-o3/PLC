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
#run '0'
#run '(letrec ([f (lambda () 5)]) 7)'
#run '(letrec ([f (lambda () 5)]) (f))'
#run '(letrec ([f (lambda () (+ 5 7))] [g (lambda () 13)]) (+ (f) (g)))'

#run '(letrec ([f (lambda (x) (+ x 1))]) (f 2))'
#run '(letrec ([f (lambda (x) (+ x 12))] [g (lambda (x) (* x (g 2)))]) (g (f 13)))'
#run '(letrec ([f (lambda (x) (+ x 1))] [g (lambda (x) (* (f x) 2))]) (g (g (g 3))))'

#run '(letrec ([f (lambda (x) (if (zero? x) x (f (sub1 x))))]) (f 0))'
#run '(letrec ([sum (lambda (n ac) (if (zero? n) ac (app sum (sub1 n) (+ n ac))))]) (sum 10 0))'

#run '(letrec ([f (lambda (x) (+ x 12))]) (f (f 10)))'

#run '(letrec ([e (lambda (x) (if (zero? x) #t (o (sub1 x))))]
#            [o (lambda (x) (if (zero? x) #f (e (sub1 x))))])
#     (e 25))'
 
#run '(letrec ([f (lambda (x acc)
#                 (if (zero? x)
#                     acc
#                     (f (sub1 x) (* acc x))))])
#      (f 5 1))'

#run '(letrec ([f (lambda (x)
#                 (if (zero? x)
#                     0
#                     (+ 1 (f (sub1 x)))))])
#			(f 200))'
#

#run '(car (car (cons (cons 12 3) (cons #t #f))))'
#run '(let ([t0 (cons 1 2)] [t1 (cons 3 4)])
#     (let ([a0 (car t0)] [a1 (car t1)] [d0 (cdr t0)] [d1 (cdr t1)])
#       (let ([t0 (cons a0 d1)] [t1 (cons a1 d0)])
#         (cons t0 t1))))'
#run '(integer? (cons 12 43))'
#run '(boolean? (cons 12 43))'
#run '(begin 12)'
#run '(let ([t (begin 13 (cons 1 2))])
#    (cons 1 t)
#    t)'


run '(let ([v (make-vector 1)] [y (cons 1 2)])
     (vector-set! v 0 y)
     (cons y (eq? y (vector-ref v 0))))'


