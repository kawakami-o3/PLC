#!/bin/sh
set -eux

goimports -w -l .
go build -o scm2asm.exe

./scm2asm.exe 32 > sandbox/scheme_entry.s


