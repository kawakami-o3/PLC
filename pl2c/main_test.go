package main

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestAdd(t *testing.T) {
	matches, _ := filepath.Glob("../test/add_*.lisp")

	for _, f := range matches {
		fmt.Println(f)
	}
}
