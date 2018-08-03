#!/bin/sh
set -eux

go-build

./scm2asm.exe 32 > sandbox/scheme_entry.s


