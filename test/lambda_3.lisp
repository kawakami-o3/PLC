(progn
	(define a 1)
	(define b 4)
	(define add (lambda (x)
								(+ x a b)))
	(print (add 3)))
