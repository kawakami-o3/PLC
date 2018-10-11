#!/usr/bin/env ruby
require "./bootstrap"

run_all "deeply nested procedures", {
  '(letrec ([e (lambda (x) (if (zero? x) #t (o (sub1 x))))]
            [o (lambda (x) (if (zero? x) #f (e (sub1 x))))])
     (e 25))' => "#f",
  '(letrec ([countdown (lambda (n)
                   (if (zero? n)
                        n
                        (countdown (sub1 n))))])
    (countdown 50005000))' => "0",
  '(letrec ([sum (lambda (n ac)
                   (if (zero? n)
                        ac
                        (sum (sub1 n) (+ n ac))))])
    (sum 10000 0))' => "50005000",
  '(letrec ([e (lambda (x) (if (zero? x) #t (o (sub1 x))))]
            [o (lambda (x) (if (zero? x) #f (e (sub1 x))))])
     (e 5000000))' => "#t"
}


