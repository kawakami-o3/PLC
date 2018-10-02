#!/usr/bin/env ruby
require "./bootstrap"

run_all "cons", {
	'(add1 0)' => "1",
  '(pair? (cons 1 2))' => "#t",
  '(pair? 12)' => "#f",
  '(pair? #t)' => "#f",
  '(pair? #f)' => "#f",
  '(pair? ())' => "#f",
  '(integer? (cons 12 43))' => "#f",
  '(boolean? (cons 12 43))' => "#f",
  '(null? (cons 12 43))' => "#f",
  '(not (cons 12 43))' => "#f",
  '(if (cons 12 43) 32 43)' => "32",
  '(car (cons 1 23))' => "1",
  '(cdr (cons 43 123))' => "123",
  '(let ((x (cons 1 2)) (y (cons 3 4))) (pair? x))' => "#t",
  '(pair? (cons (cons 12 3) #f))' => "#t",
  '(pair? (cons (cons 12 3) (cons #t #f)))' => "#t",
  '(car (car (cons (cons 12 3) (cons #t #f))))' => "12",
  '(cdr (car (cons (cons 12 3) (cons #t #f))))' => "3",
  '(car (cdr (cons (cons 12 3) (cons #t #f))))' => "#t",
  '(cdr (cdr (cons (cons 12 3) (cons #t #f))))' => "#f",
  '(pair? (cons (* 1 1) 1))' => "#t",
=begin
  '(let ([x (let ([y (+ 1 2)]) (* y y))])
     (cons x (+ x x)))' => "(10 . 18)",
  '(let ([t0 (cons 1 2)] [t1 (cons 3 4)])
     (let ([a0 (car t0)] [a1 (car t1)] [d0 (cdr t0)] [d1 (cdr t1)])
       (let ([t0 (cons a0 d1)] [t1 (cons a1 d0)])
         (cons t0 t1))))' => "((1 . 4) 3 . 2)",
=end
	'(add1 1)' => "2"
}

__END__


  [(let ([t (cons 1 2)])
     (let ([t t])
       (let ([t t])
         (let ([t t])
           t))))
   => "(1 . 2)\n"]
  [(let ([t (let ([t (let ([t (let ([t (cons 1 2)]) t)]) t)]) t)]) t)
   => "(1 . 2)\n"]
  [(let ([x ()])
     (let ([x (cons x x)])
       (let ([x (cons x x)])
         (let ([x (cons x x)])
           (cons x x)))))
   => "((((()) ()) (()) ()) ((()) ()) (()) ())\n"]
  [(cons (let ([x #t]) (let ([y (cons x x)]) (cons x y)))
         (cons (let ([x #f]) (let ([y (cons x x)]) (cons y x))) 
               ())) 
   => "((#t #t . #t) ((#f . #f) . #f))\n"]
)

(add-tests-with-string-output "begin/implicit-begin"
 [(begin 12) => "12\n"]
 [(begin 13 122) => "122\n"]
 [(begin 123 2343 #t) => "#t\n"]
 [(let ([t (begin 12 (cons 1 2))]) (begin t t)) => "(1 . 2)\n"]
 [(let ([t (begin 13 (cons 1 2))])
    (cons 1 t)
    t) => "(1 . 2)\n"]
 [(let ([t (cons 1 2)])
    (if (pair? t) 
        (begin t)
        12)) => "(1 . 2)\n"]
)

(add-tests-with-string-output "set-car! set-cdr!"
  [(let ([x (cons 1 2)])
     (begin (set-cdr! x ())
            x)) => "(1)\n"]
  [(let ([x (cons 1 2)])
     (set-cdr! x ())
     x) => "(1)\n"]
  [(let ([x (cons 12 13)] [y (cons 14 15)])
     (set-cdr! x y)
     x) => "(12 14 . 15)\n"]
  [(let ([x (cons 12 13)] [y (cons 14 15)])
     (set-cdr! y x)
     y) => "(14 12 . 13)\n"]
  [(let ([x (cons 12 13)] [y (cons 14 15)])
     (set-cdr! y x)
     x) => "(12 . 13)\n"]
  [(let ([x (cons 12 13)] [y (cons 14 15)])
     (set-cdr! x y)
     y) => "(14 . 15)\n"]
  [(let ([x (let ([x (cons 1 2)]) (set-car! x #t) (set-cdr! x #f) x)])
     (cons x x)
     x) => "(#t . #f)\n"]
  [(let ([x (cons 1 2)])
     (set-cdr! x x)
     (set-car! (cdr x) x)
     (cons (eq? x (car x)) (eq? x (cdr x)))) => "(#t . #t)\n"]
 [(let ([x #f])
    (if (pair? x)
        (set-car! x 12)
        #f)
    x) => "#f\n"]
)

(add-tests-with-string-output "more cons"
  [(letrec ([f (lambda (i lst) (if (fx= i 0) lst (f (fxsub1 i) (cons i lst))))])
(f 10 ())) => "(1 2 3 4 5 6 7 8 9 10)\n"])
