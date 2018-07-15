
(progn
	(define _eval
		(lambda (form env)
			(+ 1 1)))
	
	(define *env* (quote (nil nil)))

	(define _assoc
		(lambda (item lis)
			(if (eq lis nil)
				nil
				(if (eq (car (car lis)) item)
					(car lis)
					(_assoc item (cdr lis))))))

	(define _updatev
		(lambda (alist name value)
			(if (eq alist nil)
				nil
				(if (eq (car (car alist)) name)
					(cons (cons name (cons value nil)) (cdr alist))
					(cons (car alist) (_updatev (cdr alist) name value))))))

	(define _setv
		(lambda (name value)
			((lambda (pair)
				 (if (eq pair nil)
					 (define *env* (cons nil (cons (cons (cons name (cons value nil)) (car (cdr *env*))) nil)))
					 (define *env* (cons nil (cons (_updatev (car (cdr *env*)) name value) nil)))))
			 (_assoc name (car (cdr *env*))))))

	(define _lookup
		(lambda (name env)
			((lambda (pair)
				 (if (eq pair nil)
					 (if (eq (car env) nil)
						 (quote :unbound:)
						 (_lookup name (car env)))
					 (car (cdr pair))))
			 (_assoc name (car (cdr env))))))

	(define _bind
		(lambda (fn-args act-args env)
			(if (eq fn-args nil)
				nil
				(if (eq act-args nil)
					(quote :too-few-args:)
					(cons (cons (car fn-args) (cons (_eval (car act-args) env) nil))
								(_bind (cdr fn-args) (cdr act-args) env))))))

	(_setv (quote atom)
				 (cons (quote :lambda:)
							 (cons (quote (o))
										 (cons (lambda (ne) (atom (_lookup (quote o) ne)))
													 nil))))
	(_setv (quote cons)
				 (cons (quote :lambda:)
							 (cons (quote (a b))
										 (cons (lambda (ne) (cons (_lookup (quote a) ne) (_lookup (quote b) ne)))
													 nil))))
	(_setv (quote car)
				 (cons (quote :lambda:)
							 (cons (quote (l))
										 (cons (lambda (ne) (car (_lookup (quote l) ne)))
													 nil))))
	(_setv (quote cdr)
				 (cons (quote :lambda:)
							 (cons (quote (l))
										 (cons (lambda (ne) (cdr (_lookup (quote l) ne)))
													 nil))))
	(_setv (quote eq)
				 (cons (quote :lambda:)
							 (cons (quote (a b))
										 (cons (lambda (ne) (eq (_lookup (quote a) ne) (_lookup (quote b) ne)))
													 nil))))

	(define _eval
		(lambda (form env)
			(if (atom form)
				(if (eq form nil)
					nil
					(if (eq form t)
						form
						(_lookup form env)))
				(if (eq (cdr form) nil)
					nil
					(if (eq (car form) (quote if))
						(if (_eval (car (cdr form)) env)
							(_eval (car (cdr (cdr form))) env)
							(_eval (car (cdr (cdr (cdr form)))) env))
						(if (eq (car form) (quote quote))
							(car (cdr form))
							(if (eq (car form) (quote lambda))
								(cons (quote :lambda:)
											(cons (car (cdr form))
														(cons (lambda (ne) (_eval (car (cdr (cdr form))) ne)) nil)))
								(if (eq (car form) (quote define))
									(_setv (car (cdr form)) (_eval (car (cdr (cdr form))) env))
									((lambda (fn)
										 (if (eq fn (quote :unbound:))
											 (quote :undefined-op:)
											 ((car (cdr (cdr fn))) (cons env (cons (_bind (car (cdr fn)) (cdr form) env) nil)))))
									 (_lookup (car form) env))))))))))

	(define eval
		(lambda (form) (_eval form *env*)))

	(print (eval (quote (+ 1 2)))))
