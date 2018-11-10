

# 3.8 Procedure Calls

```
<Prog>  ::= (labels ((lvar <LExpr>) ...) <Expr>)
<LExpr> ::= (code (var ...) <Expr>)
<Expr>  ::= immediate
          | var
          | (if <Expr> <Expr> <Expr>)
          | (let ((var <Expr>) ...) <Expr)
          | (primcall prim-name <Expr> ...)
          | (labelcall lvar <Expr> ...)
```


# TODO


* new ctest, which shows how to call a dynamic label
  * 665825b5f2439c92713132f06cb7f0e671131512


## merge check list


* if? 223
* begin? 247
* make-let 262 list
* let-form? 263
* let-kind 267 car
* any-let? 268
* labels-bindings 277
* make-body 278
* let-body 282
* labels-body 285
* bind 287
* check-variable 295
* make-initial-env 299
* extend-env 301
* extend-env-with 322
* free-var 331
* free-var? 333
* free-var-offset 335 cadr
* close-env-with 337
* label? 346
* emit-variable-ref 348
* emit-any-expr [variable, closure]
* - emit-letrec
* closure-conversion 383
	* make-top list
	* top-env car
	* top-expr cadr
	* special?
	* flatmap
	* free-vars
	* emit-top
	* emit-labels
	* lambda? 'use tagged-list'
* lambda-body 467
	* make-closure
	* closure?
	* closure-label cadr
	* closure-free-vars cddr
	* emit-closre
	* make-code
	* code-formals cadr
	* code-free-variables caddr
	* code-body cadddr
	* emit-code
* app? 504
* emit-app 508
* proc 548
* define-primitive (procedure? ... 649
* emit-program 722

