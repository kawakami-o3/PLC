#include<stdio.h>
#include<stdlib.h>

typedef struct {
	int i;
} Atom;

typedef struct List {
	Atom *atom;

	struct List *car;
	struct List *cdr;
} List;

char *atom_to_string(Atom *a, char *s);
List *make_list();
List *make_int(int i);
List *car(List *lst);

List *Nil;

/*
Atom *make_atom(int i) {
	Atom *a = malloc(sizeof(Atom));
	a->i = i;
	return a;
}
*/

char *atom_to_string(Atom *a, char *s) {
	sprintf(s, "%d", a->i);
}

List *make_list() {
	List *lst = malloc(sizeof(List));
	lst->atom = NULL;
	lst->car = NULL;
	lst->cdr = NULL;
	return lst;
}

List *make_int(int i) {
	Atom *atom = malloc(sizeof(Atom));
	atom->i = i;

	List *lst = make_list();
	lst->atom = atom;
	return lst;
}

List *car(List *lst) {
	if (lst->car == NULL) {
		printf("error");
		exit(-1);
	}
	return lst->car;
}

List *cons(List *a, List *b) {
	List *lst = make_list();
	lst->car = a;
	lst->cdr = b;
	return lst;
}

List *nth(List *lst, int i) {
	if (i <= 0) {
		return lst->car;
	}
	if (lst->cdr == NULL) {
		return NULL;
	}
	return nth(lst->cdr, i-1);
}

List *add(List *lst) {
	int i = 0;

	List *a = lst->car;
	List *d = lst->cdr;
	while (a != NULL) {
		i += a->atom->i;
		a = d->car;
		d = d->cdr;
	}

	return make_int(i);
}

void init_common() {
	Nil = make_list();
}
