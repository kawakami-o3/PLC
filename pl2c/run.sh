#!/bin/sh

set -x

#TARGET=lambda_1.lisp
#TARGET=lambda_2.lisp
#TARGET=define_lambda_0.lisp
TARGET=eq_0.lisp

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

