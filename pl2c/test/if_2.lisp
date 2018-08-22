(progn
	(define fib
		(lambda (n)
			(if (eq n 0)
				1
				(if (eq n 1)
					1
					(+ (fib (- n 1)) (fib (- n 2)))))))
	(print
		(fib 0)
		(fib 1)
		(fib 2)
		(fib 3)
		(fib 4)
		(fib 5)
		(fib 6)
		(fib 7)
		(fib 8)
		(fib 9)
		(fib 10)
		))
