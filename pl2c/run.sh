set -eux
rm -f __test.c
goimports -w *.go
go build
#./purelisp.exe > test.ll
#clang test.ll
./purelisp.exe > __test.c
cat __test.c
gcc __test.c
./a.exe

