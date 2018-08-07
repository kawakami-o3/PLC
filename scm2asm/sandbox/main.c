#include<stdio.h>

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

extern int scheme_entry(void);

int main(int argc, char** argv) {
	int val = scheme_entry();

	if ((val & fixnum_mask) == fixnum_tag) {
		printf("%d\n", val >> fixnum_shift);
	} else if ((val & char_mask) == char_tag) {
		printf("#\\%c\n", val >> char_shift);
	} else if ((val & bool_mask) == bool_tag) {
		printf("%s\n", val >> bool_shift ? "#t" : "#f");
	} else if (val == empty_list) {
		printf("()\n");
	} else {
		printf("error %x\n", val);
	}

	return 0;
}
