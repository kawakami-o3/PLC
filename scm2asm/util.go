package main

func list(es []expression) expression {
	expr := expression{}
	expr.list = es
	return expr
}

func car(e expression) expression {
	return e.list[0]
}

func cadr(e expression) expression {
	return e.list[1]
}

func caddr(e expression) expression {
	return e.list[2]
}

func cdr(e expression) []expression {
	return e.list[1:]
}

func cddr(e expression) []expression {
	return e.list[2:]
}
