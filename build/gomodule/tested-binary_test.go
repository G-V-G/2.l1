package gomodule

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
)

func TestTestedBinFactory(t *testing.T) {
	ctx := blueprint.NewContext()

	ctx.MockFileSystem(map[string][]byte{
		"Blueprints": []byte(`
			go_tested_binary {
			  name: "test-out",
			  srcs: ["test-src.go"],
			  pkg: ".",
				testPkg: "./test-pkg",
  			testSrcs: ["./test-pkg/test-src_test.go"],
	      vendorFirst: true
			}
		`),
		"test-src.go": nil,
		"./test-pkg/test-src_test.go": nil,
		"out/archiveDeps.dd": nil,
	})

	ctx.RegisterModuleType("go_tested_binary", TestedBinFactory)

	cfg := bood.NewConfig()

	_, errs := ctx.ParseBlueprintsFiles(".", cfg)
	if len(errs) != 0 {
		t.Fatalf("Syntax errors in the test blueprint file: %s", errs)
	}

	_, errs = ctx.PrepareBuildActions(cfg)
	if len(errs) != 0 {
		t.Errorf("Unexpected errors while preparing build actions: %s", errs)
	}

	buffer := new(bytes.Buffer)
	if err := ctx.WriteBuildFile(buffer); err != nil {
		t.Errorf("Error writing ninja file: %s", err)
	}

	text := buffer.String()
	t.Logf("Gennerated ninja build file:\n%s", text)
	testArgs := map[string]string {
		"out/bin/test-out: ": "Generated ninja file does not have build of the test module",
		" test-src.go": "Generated ninja file does not have source dependency",
		"build vendor: g.gomodule.vendor | go.mod": "Generated ninja file does not have vendor build rule",
		"out/test/test-out.log": "Generated ninja file does not make test file log",
		"testPkg = ./test-pkg": "Generated ninja file does not have test package",
	}

	for chunk, possibleError := range testArgs {
		if !strings.Contains(text, chunk) {
			t.Errorf(possibleError)
		}
	}
}