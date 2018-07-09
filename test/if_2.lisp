(progn
	(define fib
		(lambda (n)
			(if (eq n 0)
				1
				(if (eq n 1)
					1
					(+ (fib (- n 1)) (fib (- n 2)))))))
	(print (fib 1)))
