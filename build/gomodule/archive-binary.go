package gomodule

import (
	"fmt"
	"path"

	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
)

var (
	// Ninja rule to execute go build.
  goArchive = pctx.StaticRule("archiveBuild", blueprint.RuleParams{
    Command:     "cd $workDir && zip -FSj $name $binary",
    Description: "archive binary $name",
  }, "workDir", "name", "binary", "dyndep")

	ddBuilder = pctx.StaticRule("dyndepBuild", blueprint.RuleParams{
    Command: `cd $workDir && printf 'ninja_dyndep_version = 1\nbuild $binary: dyndep\nbuild $archive: dyndep | $binary\n\n\n' > out/archiveDeps.dd`,
		Description: "create dyndep file archiveDeps.dd",
	}, "workDir", "binary", "archive")
)

// goTestedBinaryModuleType implements the simplest Go binary build with running tests for the target Go package.
type goArchiveBinaryModuleType struct {
	blueprint.SimpleName

	properties struct {
		Name string
		Binary string
		Deps []string
	}
}

func (ab *goArchiveBinaryModuleType) DynamicDependencies(blueprint.DynamicDependerModuleContext) []string {
	return ab.properties.Deps
}

func (ab *goArchiveBinaryModuleType) GenerateBuildActions(ctx blueprint.ModuleContext) {
	name := ctx.ModuleName()
	config := bood.ExtractConfig(ctx)
	config.Debug.Printf("Adding archive actions for go binary module '%s'", name)

	outputPath := path.Join(config.BaseOutputDir, "archives", ab.properties.Name)
	inputPath := path.Join(config.BaseOutputDir, "bin", ab.properties.Binary)
	ddPath := path.Join(config.BaseOutputDir, "archiveDeps.dd")

	var binaryInput []string
	if matches, err := ctx.GlobWithDeps(inputPath, make([]string, 0)); err == nil {
		binaryInput = append(binaryInput, matches...)
	} else {
		ctx.PropertyErrorf("binary", "Cannot resolve binary file %s", ab.properties.Binary)
		return
	}

	binaryInput = append(binaryInput, ddPath)

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Archive binary %s to zip %s", ab.properties.Binary, ab.properties.Name),
		Rule: goArchive,
		Outputs: []string{outputPath},
		Implicits: binaryInput,
		Args: map[string]string{
			"workDir": ctx.ModuleDir(),
			"name": outputPath,
			"binary": inputPath,
			"dyndep": ddPath,
		},
	})

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Build dyndep file"),
		Rule: ddBuilder,
		Outputs: []string{ddPath},
		Implicits: make([]string, 0),
		Args: map[string]string{
			"workDir": ctx.ModuleDir(),
			"binary": inputPath,
			"archive": outputPath,
		},
	})
}

// ArchiveBinFactory is a factory for go binary archiving.
func ArchiveBinFactory() (blueprint.Module, []interface{}) {
	mType := &goArchiveBinaryModuleType{}
	return mType, []interface{}{&mType.SimpleName.Properties, &mType.properties}
}
