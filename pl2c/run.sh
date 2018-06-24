set -x
rm -f __test.c
grep '\tcode' main.go
goimports -w *.go
go build -o __pl2c.exe && ./__pl2c.exe > __test.c && cat __test.c

#go build -o __pl2c.exe && ./__pl2c.exe
#go build -o __pl2c.exe
#./purelisp.exe > test.ll
#clang test.ll



gcc __test.c -o __a.exe && ./__a.exe

