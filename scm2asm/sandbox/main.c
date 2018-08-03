#include<stdio.h>

extern int scheme_entry(void);

int main(int argc, char** argv) {
	printf("%d\n", scheme_entry());
	return 0;
}
