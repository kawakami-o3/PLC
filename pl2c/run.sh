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
TARGET=quote_8.lisp

rm -f __test.c
grep '\tcode' main.go
goimports -w *.go
#go build -o __pl2c.exe && ./__pl2c.exe > __test.c && cat __test.c

#go build -o __pl2c.exe && ./__pl2c.exe ../test/add_0.lisp > __test.c && cat __test.c
go build -o __pl2c.exe && ./__pl2c.exe ../test/$TARGET > __test.c && cat __test.c
#go build -o __pl2c.exe && ./__pl2c.exe ../test/$TARGET

#go build -o __pl2c.exe && ./__pl2c.exe
#go build -o __pl2c.exe
#./purelisp.exe > test.ll
#clang test.ll



gcc __test.c -o __a.exe && ./__a.exe

