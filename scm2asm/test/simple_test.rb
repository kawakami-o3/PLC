#!/usr/bin/env ruby
require "./bootstrap"

run 0
run 42
run '#\a'
run '#\z'
run '#\M'
run '#t'
run '#f'
run '()'
run '(add1 3)', 4
run '(sub1 3)', 2
run '(integer->char 65)', '#\A'
run '(char->integer #\A)', '65'
run '(zero? 1)', '#f'
run '(zero? 0)', '#t'
run '(null? ())', '#t'
run '(null? 1)', '#f'
run '(not #t)', '#f'
run '(not #f)', '#t'

run '(integer? 1)', '#t'
run '(integer? #\a)', '#f'
run '(integer? #t)', '#f'
run '(boolean? #t)', '#t'
run '(boolean? 8)', '#f'
run '(boolean? #\c)', '#f'

run '(+ 3 10)', 13
run '(- 3 10)', -7
run '(* 3 10)', 30
run '(= 10 10)', '#t'
run '(= 3 10)', '#f'
run '(< 10 1)', '#f'
run '(< 1 10)', '#t'
run '(<= 1 10)', '#t'
run '(<= 10 10)', '#t'
run '(<= 11 10)', '#f'
run '(> 1 10)', '#f'
run '(> 10 1)', '#t'
run '(>= 1 10)', '#f'
run '(>= 10 10)', '#t'
run '(>= 11 10)', '#t'
run '(char=? #\a #\a)', '#t'
run '(char=? #\a #\c)', '#f'
run '(let ((a 1)) a)', 1
run '(let ((a 1)) (+ a 3))', 4
run '(let* ((a 2) (b (+ a 3))) (* a b))', 10
run '(if (= 1 1) 10 20)', 10
run '(if (= 1 2) 10 20)', 20
run '(if (= 1 2) 10 (+ 20 3))', 23

#run '(car (cons 1 2))', 10


