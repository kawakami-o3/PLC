#!/usr/bin/env ruby

require "./bootstrap"

run_all "procedure?", {
  '(procedure? (lambda (x) x))' => "#t\n",
  '(let ([f (lambda (x) x)]) (procedure? f))' => "#t\n",
  '(procedure? (make-vector 0))' => "#f\n",
  '(procedure? (make-string 0))' => "#f\n",
  '(procedure? (cons 1 2))' => "#f\n",
  '(procedure? #\S)' => "#f\n",
  '(procedure? ())' => "#f\n",
  '(procedure? #t)' => "#f\n",
  '(procedure? #f)' => "#f\n",
  '(string? (lambda (x) x))' => "#f\n",
  '(vector? (lambda (x) x))' => "#f\n",
  '(boolean? (lambda (x) x))' => "#f\n",
  '(null? (lambda (x) x))' => "#f\n",
  '(not (lambda (x) x))' => "#f\n"
}



__END__

(add-tests-with-string-output "applying thunks"
  [(let ([f (lambda () 12)]) (f)) => "12\n"]
  [(let ([f (lambda () (fx+ 12 13))]) (f)) => "25\n"]
  [(let ([f (lambda () 13)]) (fx+ (f) (f))) => "26\n"]
  [(let ([f (lambda () 
              (let ([g (lambda () (fx+ 2 3))])
                (fx* (g) (g))))])
    (fx+ (f) (f))) => "50\n"]
  [(let ([f (lambda () 
              (let ([f (lambda () (fx+ 2 3))])
                (fx* (f) (f))))])
    (fx+ (f) (f))) => "50\n"]
  [(let ([f (if (boolean? (lambda () 12))
                (lambda () 13)
                (lambda () 14))])
     (f)) => "14\n"]
)


(add-tests-with-string-output "parameter passing"
 [(let ([f (lambda (x) x)]) (f 12)) => "12\n"]
 [(let ([f (lambda (x y) (fx+ x y))]) (f 12 13)) => "25\n"]
 [(let ([f (lambda (x)
             (let ([g (lambda (x y) (fx+ x y))])
               (g x 100)))])
   (f 1000)) => "1100\n"]
 [(let ([f (lambda (g) (g 2 13))])
    (f (lambda (n m) (fx* n m)))) => "26\n"]
 [(let ([f (lambda (g) (fx+ (g 10) (g 100)))])
   (f (lambda (x) (fx* x x)))) => "10100\n"]
 [(let ([f (lambda (f n m)
             (if (fxzero? n)
                 m
                 (f f (fxsub1 n) (fx* n m))))])
   (f f 5 1)) => "120\n"]
 [(let ([f (lambda (f n)
             (if (fxzero? n)
                 1
                 (fx* n (f f (fxsub1 n)))))])
   (f f 5)) => "120\n"]
)
 

(add-tests-with-string-output "closures"
 [(let ([n 12])
    (let ([f (lambda () n)])
      (f))) => "12\n"]
 [(let ([n 12])
    (let ([f (lambda (m) (fx+ n m))])
      (f 100))) => "112\n"]
 [(let ([f (lambda (f n m)
             (if (fxzero? n)
                 m
                 (f (fxsub1 n) (fx* n m))))])
   (let ([g (lambda (g n m) (f (lambda (n m) (g g n m)) n m))])
     (g g 5 1))) => "120\n"]
 [(let ([f (lambda (f n)
             (if (fxzero? n)
                 1
                 (fx* n (f (fxsub1 n)))))])
   (let ([g (lambda (g n) (f (lambda (n) (g g n)) n))])
     (g g 5))) => "120\n"]
)
