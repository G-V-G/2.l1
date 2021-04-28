package gomodule

import (
	"fmt"
	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
	"path"
)

var (
	// Ninja rule to execute go build.
	goBuild = pctx.StaticRule("binaryBuild", blueprint.RuleParams{
		Command:     "cd $workDir && go build -o $outputPath $pkg",
		Description: "build go command $pkg",
	}, "workDir", "outputPath", "pkg")

	// Ninja rule to execute go mod vendor.
	goVendor = pctx.StaticRule("vendor", blueprint.RuleParams{
		Command:     "cd $workDir && go mod vendor",
		Description: "vendor dependencies of $name",
	}, "workDir", "name")

	goTest = pctx.StaticRule("test", blueprint.RuleParams{
		Command: "cd $workDir && go test $testPkg > $outputPath",
		Description: "test package $testPkg",
	}, "workDir", "testPkg", "outputPath")
)

// goTestedBinaryModuleType implements the simplest Go binary build with running tests for the target Go package.
type goTestedBinaryModuleType struct {
	blueprint.SimpleName

	properties struct {
		// Go package for testing.
		TestPkg string
		// Test files template.
		TestSrcs []string
		// Go package name to build as a command with "go build".
		Pkg string
		// List of source files.
		Srcs []string
		// Exclude patterns.
		SrcsExclude []string
		// If to call vendor command.
		VendorFirst bool
		// Optional execution.
		OptionalBuild bool
		OptionalTest bool

		// Example of how to specify dependencies.
		Deps []string
	}
}

func (gb *goTestedBinaryModuleType) DynamicDependencies(blueprint.DynamicDependerModuleContext) []string {
	return gb.properties.Deps
}

func (gb *goTestedBinaryModuleType) GenerateBuildActions(ctx blueprint.ModuleContext) {
	name := ctx.ModuleName()
	config := bood.ExtractConfig(ctx)
	config.Debug.Printf("Adding build actions for go binary module '%s'", name)

	outputPath := path.Join(config.BaseOutputDir, "bin", name)
	outputTestPath := path.Join(config.BaseOutputDir, "test", name + ".log")

	var inputs []string
	inputErors := false
	for _, src := range gb.properties.Srcs {
		if matches, err := ctx.GlobWithDeps(src, append(gb.properties.SrcsExclude, gb.properties.TestSrcs...)); err == nil {
			inputs = append(inputs, matches...)
		} else {
			ctx.PropertyErrorf("srcs", "Cannot resolve files that match pattern %s", src)
			inputErors = true
		}
	}

	var testInputs []string
	for _, src := range gb.properties.TestSrcs {
		if matches, err := ctx.GlobWithDeps(src, make([]string, 0)); err == nil {
			testInputs = append(testInputs, matches...)
		} else {
			ctx.PropertyErrorf("srcs", "Cannot resolve files that match pattern %s", src)
			inputErors = true
		}
	}
	if inputErors {
		return
	}

	if gb.properties.VendorFirst {
		vendorDirPath := path.Join(ctx.ModuleDir(), "vendor")
		ctx.Build(pctx, blueprint.BuildParams{
			Description: fmt.Sprintf("Vendor dependencies of %s", name),
			Rule:        goVendor,
			Outputs:     []string{vendorDirPath},
			Implicits:   []string{path.Join(ctx.ModuleDir(), "go.mod")},
			Optional:    true,
			Args: map[string]string{
				"workDir": ctx.ModuleDir(),
				"name":    name,
			},
		})
		inputs = append(inputs, vendorDirPath)
	}

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Execute and write tests for %s package", name),
		Rule: goTest,
		Outputs: []string{outputTestPath},
		Implicits: append(testInputs, inputs...),
		Optional: gb.properties.OptionalTest,
		Args: map[string]string{
			"workDir": ctx.ModuleDir(),
			"testPkg": gb.properties.TestPkg,
			"outputPath": outputTestPath,
		},
	})

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Build %s as Go binary", name),
		Rule:        goBuild,
		Outputs:     []string{outputPath},
		Implicits:   append(inputs),
		Optional: gb.properties.OptionalBuild,
		Args: map[string]string{
			"outputPath": outputPath,
			"workDir":    ctx.ModuleDir(),
			"pkg":        gb.properties.Pkg,
		},
	})	
}

// TestedBinFactory is a factory for go binary module type which supports Go command packages with running tests.
func TestedBinFactory() (blueprint.Module, []interface{}) {
	mType := &goTestedBinaryModuleType{}
	return mType, []interface{}{&mType.SimpleName.Properties, &mType.properties}
}