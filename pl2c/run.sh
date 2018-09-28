#!/bin/sh

set -x

#TARGET=lambda_1.lisp
#TARGET=lambda_2.lisp
#TARGET=define_lambda_0.lisp
#TARGET=eq_0.lisp
#TARGET=eq_1.lisp
#TARGET=eq_2.lisp
#TARGET=if_0.lisp
#TARGET=if_1.lisp
#TARGET=if_2.lisp
#TARGET=quote_0.lisp
#TARGET=quote_1.lisp
#TARGET=quote_2.lisp
#TARGET=quote_3.lisp
#TARGET=quote_4.lisp
#TARGET=quote_5.lisp
#TARGET=quote_6.lisp
#TARGET=atom_0.lisp
#TARGET=atom_1.lisp

# to eval
#TARGET=quote_7.lisp
#TARGET=quote_8.lisp
#TARGET=define_0.lisp
#TARGET=define_1.lisp
#TARGET=define_2.lisp
#TARGET=nil_0.lisp
TARGET=lambda_3.lisp
#TARGET=eval.lisp

make clean
make

go build -o pl2c.exe && ./pl2c.exe ./test/$TARGET
#go build -o pl2c.exe && ./pl2c.exe -S -o ./test.c ./test/$TARGET


