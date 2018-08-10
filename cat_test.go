package main

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

var (
	testUpdateGoldFile = false
)

func init() {
	flag.BoolVar(&testUpdateGoldFile, "update", false, "update gold files")
}

type subCmdFunc func(*cobra.Command, []string) error

func withInput(t *testing.T, fn string, f subCmdFunc) subCmdFunc {
	t.Helper()
	in, err := os.Open(fn)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	return func(cmd *cobra.Command, args []string) error {
		os.Stdin = in
		defer in.Close()
		return f(cmd, args)
	}
}

func runSubCmd(t *testing.T, f subCmdFunc, args ...string) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	defer func() { _ = r.Close() }()
	os.Stdout = w
	if err = f(nil, args); err != nil {
		t.Fatalf("got error: %v", err)
	}
	if err = w.Close(); err != nil {
		t.Fatalf("got error: %v", err)
	}
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	os.Stdout = oldStdout
	return string(bs)
}

func checkGoldFile(t *testing.T, gold, got string) {
	// update the gold file with the given output
	if testUpdateGoldFile {
		if err := ioutil.WriteFile(gold, []byte(got), os.ModePerm); err != nil {
			t.Fatalf("got error: %v", err)
		}
	}
	t.Helper()
	in, err := os.Open(gold)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	defer func() { _ = in.Close() }()
	want, err := ioutil.ReadAll(in)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if string(want) != got {
		t.Fatalf("expected %q; got %q in %s", want, got, gold)
	}
}

func TestCatCmd(t *testing.T) {
	got := runSubCmd(t, cat, "testdata/0001.gt.txt", "testdata/0002.gt.txt")
	checkGoldFile(t, "testdata/cat_output_gold.txt", got)
}

func TestAlignCmd(t *testing.T) {
	got := runSubCmd(t, withInput(t, "testdata/cat_output_gold.txt", align))
	checkGoldFile(t, "testdata/align_output_gold.txt", got)
}
