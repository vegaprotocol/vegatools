package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer w.Close()
	f()
	os.Stdout = old
	w.Close()
	out, _ := ioutil.ReadAll(r)
	return string(out)
}

func TestMain(t *testing.T) {
	output := captureOutput(func() {
		main()
	})
	assert.Contains(t, output, "Usage")
}
