package main

import (
	"os"
	"testing"
)

func TestMain_Run(t *testing.T) {
	t.Parallel()
	os.Args = []string{"cmd"}
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("main panicked: %v", r)
		}
	}()
	main()
}
