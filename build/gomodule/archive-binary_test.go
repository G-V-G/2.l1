package gomodule

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
)

func TestArchiveBinFactory(t *testing.T) {
	ctx := blueprint.NewContext()

	ctx.MockFileSystem(map[string][]byte{
		"Blueprints": []byte(`
			archive_bin {
				name: "test-archive",
				binary: "test-out"
			}
		`),
		"out/archiveDeps.dd": nil,
	})

	ctx.RegisterModuleType("archive_bin", ArchiveBinFactory)

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
		"build out/archives/test-archive:": "Generated ninja file does not have build of the expected archive",
		"binary = out/bin/test-out": "Generated ninja file does not have specified binary name",
		"name = out/archives/test-archive": "Generated ninja file does not have specified archive name",
	}

	for chunk, possibleError := range testArgs {
		if !strings.Contains(text, chunk) {
			t.Errorf(possibleError)
		}
	}
}
