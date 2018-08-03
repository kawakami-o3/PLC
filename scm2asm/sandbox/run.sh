#!/bin/sh

set -eux

#gcc -c scheme_entry.s
#gcc -S scheme_entry.o main.c
gcc -o a.out scheme_entry.s main.c

./a.out
