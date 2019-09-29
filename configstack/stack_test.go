package configstack

import (
	"github.com/gruntwork-io/terragrunt/config"
	"github.com/gruntwork-io/terragrunt/options"
	"github.com/gruntwork-io/terragrunt/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindStackInSubfolders(t *testing.T) {
	t.Parallel()

	filePaths := []string{
		"/stage/data-stores/redis/" + config.DefaultTerragruntConfigPath,
		"/stage/data-stores/postgres/" + config.DefaultTerragruntConfigPath,
		"/stage/ecs-cluster/" + config.DefaultTerragruntConfigPath,
		"/stage/kms-master-key/" + config.DefaultTerragruntConfigPath,
		"/stage/vpc/" + config.DefaultTerragruntConfigPath,
	}

	tempFolder := createTempFolder(t)
	writeDummyTerragruntConfigs(t, tempFolder, filePaths)

	envFolder := filepath.ToSlash(util.JoinPath(tempFolder + "/stage"))
	terragruntOptions, err := options.NewTerragruntOptions(envFolder)
	if err != nil {
		t.Fatalf("Failed when calling method under test: %s\n", err.Error())
	}

	terragruntOptions.WorkingDir = envFolder

	stack, err := FindStackInSubfolders(terragruntOptions)
	require.NoError(t, err)

	var modulePaths []string

	for _, module := range stack.Modules {
		relPath := strings.Replace(module.Path, tempFolder, "", 1)
		relPath = filepath.ToSlash(util.JoinPath(relPath, config.DefaultTerragruntConfigPath))

		modulePaths = append(modulePaths, relPath)
	}

	for _, filePath := range filePaths {
		filePathFound := util.ListContainsElement(modulePaths, filePath)
		assert.True(t, filePathFound, "The filePath %s was not found by Terragrunt.\n", filePath)
	}

}

func createTempFolder(t *testing.T) string {
	tmpFolder, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %s\n", err.Error())
	}

	return filepath.ToSlash(tmpFolder)
}

// Create a dummy Terragrunt config file at each of the given paths
func writeDummyTerragruntConfigs(t *testing.T, tmpFolder string, paths []string) {
	contents := []byte("terraform {\nsource = \"test\"\n}\n")
	for _, path := range paths {
		absPath := util.JoinPath(tmpFolder, path)

		containingDir := filepath.Dir(absPath)
		createDirIfNotExist(t, containingDir)

		err := ioutil.WriteFile(absPath, contents, os.ModePerm)
		if err != nil {
			t.Fatalf("Failed to write file at path %s: %s\n", path, err.Error())
		}
	}
}

func createDirIfNotExist(t *testing.T, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			t.Fatalf("Failed to create directory: %s\n", err.Error())
		}
	}
}
