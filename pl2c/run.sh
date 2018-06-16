set -eux
rm -f __test.c
goimports -w *.go
go build -o __pl2c.exe
#./purelisp.exe > test.ll
#clang test.ll
./__pl2c.exe > __test.c
cat __test.c


gcc __test.c -o __a.exe && ./__a.exe

