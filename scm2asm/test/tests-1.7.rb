#!/usr/bin/env ruby
require "./bootstrap"

run '(letrec () 12)', 12
run '(letrec () (let ([x 5]) (+ x x)))', 10
run '(letrec ([f (lambda () 5)]) 7)', 7
run '(letrec ([f (lambda () 5)]) (let ([x 12]) x))', 12
run '(letrec ([f (lambda () 5)]) (f))', 5
run '(letrec ([f (lambda () 5)]) (let ([x (f)]) x))', 5
run '(letrec ([f (lambda () 5)]) (+ (f) 6))', 11
run '(letrec ([f (lambda () 5)]) (+ 6 (f)))', 11
run '(letrec ([f (lambda () 5)]) (- 20 (f)))', 15
run '(letrec ([f (lambda () 5)]) (+ (f) (f)))', 10
run '(letrec ([f (lambda () (+ 5 7))] [g (lambda () 13)]) (+ (f) (g)))', 25
run '(letrec ([f (lambda (x) (+ x 12))]) (f 13))', 25
run '(letrec ([f (lambda (x) (+ x 12))]) (f (f 10)))', 34
run '(letrec ([f (lambda (x) (+ x 12))]) (f (f (f 0))))', 36
run '(letrec ([f (lambda (x y) (+ x y))]
            [g (lambda (x) (+ x 12))])
    (f 16 (f (g 0) (+ 1 (g 0)))))', 41
run '(letrec ([f (lambda (x) (g x x))]
            [g (lambda (x y) (+ x y))])
     (f 12))', 24

run '(letrec ([f (lambda (x)
                 (if (zero? x)
                     1
                     (* x (f (sub1 x)))))])
      (f 5))', 120

run '(letrec ([f (lambda (x acc)
                 (if (zero? x)
                     acc
                     (f (sub1 x) (* acc x))))])
      (f 5 1))', 120


run '(letrec ([e (lambda (x) (if (zero? x) #t (o (sub1 x))))]
            [o (lambda (x) (if (zero? x) #f (e (sub1 x))))])
     (e 5))', "#f"

__END__
run '(letrec ([f (lambda (x)
                 (if (zero? x)
                     0
                     (+ 1 (f (sub1 x)))))])
			(f 200))', 200




(add-tests-with-string-output "more stack"
  [(letrec ([f (lambda (n)
                 (if (fxzero? n)
                     0
                     (fx+ 1 (f (fxsub1 n)))))])
(f 500)) => "500\n"])

=begin
run '(letrec ([f (lambda (x)
                 (if (zero? x)
                     0
                     (+ 1 (f (sub1 x)))))])
      (f 200))', 200
=end


def hoge v
	run "(letrec ([f (lambda (x)
									 (if (zero? x)
											 0
											 (+ 1 (f (sub1 x)))))])
				(f #{v}))", v
end

(1..16).each do |i|
	hoge i
end


