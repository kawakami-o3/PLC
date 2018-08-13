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

typedef unsigned int ptr;

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


static void print_ptr(ptr x) {
	if ((x & fixnum_mask) == fixnum_tag) {
		printf("%d\n", ((int)x) >> fixnum_shift);
	} else if ((x & char_mask) == char_tag) {
		printf("#\\%c\n", x >> char_shift);
	} else if ((x & bool_mask) == bool_tag) {
		printf("%s\n", x >> bool_shift ? "#t" : "#f");
	} else if (x == empty_list) {
		printf("()\n");
	} else {
		printf("error %x\n", x);
	}
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

extern ptr scheme_entry(char* stack_base);
//extern ptr scheme_entry(context* ctxt, char* stack_base, memory* mem);

int main(int argc, char** argv) {
	int stack_size = (16 * 4096);
	char* stack_top = allocate_protected_space(stack_size);
	char* stack_base = stack_top + stack_size;

	print_ptr(scheme_entry(stack_base));

	deallocate_protected_space(stack_top, stack_size);
	return 0;
}
