#include<stdarg.h>
#include<stdio.h>
#include<stdlib.h>
#include<string.h>

#define INIT_SIZE 8

void printLine() {
	printf("%d\n", __LINE__);
}

typedef struct {
	int nalloc;
	int length;
	char *chars;
} String;

String *make_string() {
	String *s = malloc(sizeof(String));
	s->chars = malloc(INIT_SIZE);
	s->nalloc = INIT_SIZE;
	s->length = 0;
	s->chars[0] = '\0';
	return s;
}

static void realloc_string(String *s) {
	int newsize = s->nalloc * 2;
	char *chars = malloc(newsize);
	strcpy(chars, s->chars);
	s->chars = chars;
	s->nalloc = newsize;
}

void string_append(String *s, char c) {
	if (s->nalloc == (s->length + 1)) {
		realloc_string(s);
	}
	s->chars[s->length++] = c;
	s->chars[s->length] = '\0';
}

void string_appendf(String *s, char *fmt, ...) {
	va_list args;
	for (;;) {
		int avail = s->nalloc - s->length;
		va_start(args, fmt);
		int written = vsnprintf(s->chars + s->length, avail, fmt, args);
		va_end(args);
		if (avail <= written) {
			realloc_string(s);
			continue;
		}
		s->length += written;
		return;
	}
}

enum Type {
	NIL,
	INT,
	STR,
	PROC
};

struct List;

typedef struct {
	int type;

	int i;
	String *str;
	struct List *(*proc)(struct List*);
} Atom;

typedef struct List {
	Atom *atom;

	struct List *car;
	struct List *cdr;
} List;

List *make_list();
List *make_int(int i);
List *car(List *lst);

List *nil;
List *t;

Atom *make_atom() {
	Atom *a = malloc(sizeof(Atom));
	return a;
}

List *make_list() {
	List *lst = malloc(sizeof(List));
	lst->atom = NULL;
	lst->car = NULL;
	lst->cdr = NULL;
	return lst;
}

List *make_int(int i) {
	Atom *atom = make_atom();
	atom->type = INT;
	atom->i = i;

	List *lst = make_list();
	lst->atom = atom;
	return lst;
}

List *make_symbol(char *chars) {
	Atom *atom = make_atom();
	atom->type = STR;
	atom->str = make_string();
	string_appendf(atom->str, "%s", chars);

	List *lst = make_list();
	lst->atom = atom;
	return lst;
}

List *make_lambda(List *(*proc)(struct List*)) {
	Atom *atom = make_atom();
	atom->type = PROC;
	atom->proc = proc;

	List *lst = make_list();
	lst->atom = atom;
	return lst;
}

List *eq(List *a, List *b) {
	if (a->atom != NULL && b->atom != NULL) {
		if (a->atom->type != b->atom->type) {
			return nil;
		} else if (a->atom->type == INT) {
			return a->atom->i == b->atom->i ? t : nil;
		} else if (a->atom->type == STR) {
			return strcmp(a->atom->str->chars, b->atom->str->chars) == 0 ? t : nil;
		} else if (a->atom->type == PROC) {
			return a->atom->proc == b->atom->proc ? t : nil;
		}
	} else if (a->car != NULL && b->car != NULL) {
		if (eq(a->car, b->car)) {
			if (a->cdr != NULL && b->cdr != NULL) {
				return eq(a->cdr, b->cdr);
			} else if (a->cdr == NULL && b->cdr == NULL) {
				return t;
			} else {
				return nil;
			}
			return nil;
		} else {
			return nil;
		}
	} else {
		return nil;
	}
}

void to_string(String *str, List *lst) {
	if (lst == NULL) {
		return;
	} else if (lst->atom == NULL) {
		string_appendf(str, "(");
		to_string(str, lst->car);
		string_appendf(str, " ");
		if (lst->cdr != NULL && lst->cdr != nil) {
			to_string(str, lst->cdr);
		}
		string_appendf(str, ")");
	} else {
		switch(lst->atom->type) {
			case INT:
				string_appendf(str, " %d", lst->atom->i);
				break;
			case STR:
				string_appendf(str, " %s", lst->atom->str->chars);
				break;
			case PROC:
				string_appendf(str, " PROC");
				break;
			case NIL:
				string_appendf(str, " nil");
				break;
			default:
				string_appendf(str, " ???");
		}
	}
}

/*
List *callProc(List *proc, int argc, ...) {
	va_list args;
	List *ret;
	List *lst = make_list();
	va_start(args, argc);
	for (int i=0 ; i<argc ; i++) {
		List *a = va_arg(args, (List*));
		lst = cons(a, lst);
	}
	va_end(args);
	return proc.proc(lst);
}
*/

void printList(List *lst) {
	String *s = make_string();
	to_string(s, lst);
	printf("%s\n", s->chars);
}

List *car(List *lst) {
	if (lst->car == NULL) {
		printf("error");
		exit(-1);
	}
	return lst->car;
}

List *cdr(List *lst) {
	return lst->cdr;
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

List *plc_add(List *lst) {
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

List *plc_sub(List *lst) {
	List *a = lst->car;
	List *d = lst->cdr;
	int i = a->atom->i;

	a = d->car;
	d = d->cdr;
	while (a != NULL) {
		i -= a->atom->i;
		a = d->car;
		d = d->cdr;
	}

	return make_int(i);
}

void init_common() {
	nil = make_list();
	nil->atom = make_atom();
	nil->atom->type = NIL;
	nil->atom->i = 0;

	t = make_list();
	t->atom = make_atom();
	nil->atom->type = INT;
	t->atom->i = 1;
}
