package main

import (
	"bytes"
	"testing"
)

func TestVersion(t *testing.T) {
	var output bytes.Buffer
	if err := run([]string{"--version"}, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if output.String() != version+"\n" {
		t.Fatalf("unexpected version output: %q", output.String())
	}
}
