package main

import (
	"os"
	"testing"
)

func TestMainFunction(t *testing.T) {
	orig := os.Args
	defer func() { os.Args = orig }()
	os.Args = []string{"skcr", "version"}
	main()
}
