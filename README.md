# PLC
poorman's lisp compiler

# Installation

```
go get -u github.com/kawakami-o3/PLC/...
```

# Example

```
% echo '(print (+ 1 2 3))' > test.lisp
% pl2c test.lisp
% a.out
 6
```
