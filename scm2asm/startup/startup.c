#include<stdio.h>
#include<stdlib.h>
#include<sys/mman.h>
#include<unistd.h>

#define fixnum_mask  3 // 11
#define fixnum_tag   0 // 00
#define fixnum_shift 2

#define char_mask  0xFF // 11111111
#define char_tag   0x0F // 00001111
#define char_shift 8

#define bool_mask  0x7F // 01111111
#define bool_tag   0x1F // 00011111
#define bool_shift 7

#define empty_list 0x2F // 00101111

#define obj_mask  0x07
#define obj_shift 8

#define pair_tag  0x01
#define pair_size 16
#define pair_car  0
#define pair_cdr  8

#define vector_tag 0x05
#define string_tag 0x06
#define closure_tag 0x02

typedef unsigned long ptr;

typedef struct {
    void* eax; /* 0  scratch  */
    void* ebx; /* 4  preserve */
    void* ecx; /* 8  scratch  */
    void* edx; /* 12 scratch  */
    void* esi; /* 16 preserve */
    void* edi; /* 20 preserve */
    void* ebp; /* 24 preserve */
    void* esp; /* 28 preserve */
} context;

typedef struct {
    ptr car;
    ptr cdr;
} cell;

typedef struct {
    ptr length;
    ptr buf[1];
} vector;

typedef struct {
    ptr length;
    char buf[1];
} string;

typedef struct {
  char* heap_next;
  char* global_next;
  ptr edi;
  char* heap_base;
  char* heap_top;
  char* heap_base_alt;
  char* heap_top_alt;
  char* global_base;
  char* stack_base;
  char* scratch_base;
} memory;

#define IN_LIST 1
#define OUT 0

static void print_ptr_rec(ptr x, int state) {
	if ((x & fixnum_mask) == fixnum_tag) {
		printf("%d", ((int)x) >> fixnum_shift);
	} else if ((x & bool_mask) == bool_tag) {
		printf("%s", x >> bool_shift ? "#t" : "#f");
	} else if (x == empty_list) {
		printf("()");
    } else if ((x & char_mask) == char_tag) {
        char c = (char) (x >> char_shift);
        if (c == '\t') {
            printf("#\\tab");
        } else if (c == '\n') {
            printf("#\\newline");
        } else if (c == '\r') {
            printf("#\\return");
        } else if (c == ' ') {
            printf("#\\space");
        } else {
            printf("#\\%c", c);
        }
    } else if ((x & obj_mask) == pair_tag) {
        if (state != IN_LIST) {
            printf("(");
        }
        ptr car = ((cell*)(x-pair_tag))->car;
        print_ptr_rec(car, OUT);
        ptr cdr = ((cell*)(x-pair_tag))->cdr;
        if (cdr != empty_list) {
            if ((cdr & obj_mask) == pair_tag) {
                printf(" ");
                print_ptr_rec(cdr, IN_LIST);
            } else {
                printf(" . ");
                print_ptr_rec(cdr, OUT);
            }
        }
        if (state != IN_LIST) {
            printf(")");
        }
    } else if ((x & obj_mask) == vector_tag) {
        printf("#(");

        vector* p = (vector*)(x-vector_tag);
        unsigned long n = p->length >> fixnum_shift;
        unsigned long i;
        for (i=0 ; i<n ; i++) {
            if (i>0) {
                printf(" ");
            }
            print_ptr_rec(p->buf[i], OUT);
        }

        printf(")");
    } else if ((x & obj_mask) == string_tag) {
        printf("\"");

        string* p = (string*)(x-string_tag);
        unsigned long n = p->length >> fixnum_shift;
        unsigned long i;
        for (i=0 ; i<n ; i++) {
            int c = p->buf[i];

            if (c == '"') {
                printf("\\\"");
            } else if (c == '\\') {
                printf("\\\\");
            } else {
                putchar(c);
            }
        }

        printf("\"");
	} else if ((x & obj_mask) == closure_tag) {
		printf("#<procedure>");
	} else {
		printf("#<unknown 0x%08lx>", x);
	}
}

static void print_ptr(ptr x) {
    print_ptr_rec(x, OUT);
    printf("\n");
}

static char* allocate_protected_space(int size) {
	int page = getpagesize();
	int status;
	int aligned_size = ((size + page - 1) / page) * page;
	char* p = mmap(0, aligned_size + 2 * page,
			PROT_READ | PROT_WRITE,
			MAP_ANONYMOUS | MAP_PRIVATE,
			0, 0);
	if (p == MAP_FAILED) {
		perror("map");
		exit(1);
	}
	status = mprotect(p, page, PROT_NONE);
	if (status != 0) {
		perror("mprotect");
		exit(status);
	}
	status = mprotect(p + page + aligned_size, page, PROT_NONE);
	if (status != 0) {
		perror("mprotect");
		exit(status);
	}
	return (p + page);
}

static void deallocate_protected_space(char* p, int size) {
	int page = getpagesize();
	int status;
	int aligned_size = ((size + page - 1) / page) * page;
	status = munmap(p - page, aligned_size + 2 * page);
	if (status != 0) {
		perror("munmap");
		exit(status);
	}
}

//extern ptr scheme_entry(char* stack_base);
//extern ptr scheme_entry(context* ctxt, char* stack_base, memory* mem);
extern ptr scheme_entry(context* ctxt, char* stack_base, char* heap);

int main(int argc, char** argv) {
	int stack_size = (16 * 4096);
	int heap_size = (16 * 4096);

	char* stack_top = allocate_protected_space(stack_size);
	char* stack_base = stack_top + stack_size;

    char* heap = allocate_protected_space(heap_size);

    context ctxt;
	//printf("%d\n", stack_top);
	print_ptr(scheme_entry(&ctxt, stack_base, heap));
	//printf("%d\n", stack_top);

	deallocate_protected_space(stack_top, stack_size);
	deallocate_protected_space(heap, heap_size);
	return 0;
}
