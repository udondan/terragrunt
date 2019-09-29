package configstack

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gruntwork-io/terragrunt/errors"
	"github.com/gruntwork-io/terragrunt/options"
	"github.com/gruntwork-io/terragrunt/util"
	"github.com/kr/pretty"
)

const tmpStatefileName = "terragrunt-mv-tmp.tfstate"
const moveManifestName = ".terragrunt-mv-manifest"

// MvAll moves a local directory/tree and the related statefiles to the matching remote location
func MvAll(terragruntOptions *options.TerragruntOptions) error {
	stackOrigin, err := FindStackInSubfolders(terragruntOptions)
	if err != nil {
		return err
	}

	if err := stackOrigin.StatePull(terragruntOptions, tmpStatefileName); err != nil {
		return err
	}

	defer cleanTmpFiles(stackOrigin, tmpStatefileName)

	terragruntOptionsTarget, err := newOptionsWithReplacedPath(terragruntOptions)
	if err != nil {
		return err
	}

	if err := util.CopyFolderContents(terragruntOptions.WorkingDir, terragruntOptionsTarget.WorkingDir, moveManifestName, false); err != nil {
		return err
	}

	stackTarget, err := FindStackInSubfolders(terragruntOptionsTarget)
	if err != nil {
		return err
	}

	// walk through stack and exclude modules w/o tmp statefile
	var filteredModules []*TerraformModule
	for _, module := range stackTarget.Modules {
		stateFile := path.Join(module.Path, tmpStatefileName)
		if util.FileExists(stateFile) {
			filteredModules = append(filteredModules, module)
			//module.Config.Skip = true // this doesn't do anything
		}
	}
	stackTarget.Modules = filteredModules

	if err := stackTarget.StatePush(terragruntOptionsTarget, tmpStatefileName); err != nil {
		return err
	}
	/*
		if err := stackOrigin.StateRm(terragruntOptions); err != nil {
			return err
		}
		/**/

	if err := cleanTmpFiles(stackOrigin, tmpStatefileName); err != nil {
		return err
	}
	if err := cleanTmpFiles(stackTarget, tmpStatefileName); err != nil {
		return err
	}

	//return os.RemoveAll(terragruntOptions.WorkingDir)
	return nil

}

// replacePathsInOptions updates the keys holding paths, replace the old working directory with the new working directory
func newOptionsWithReplacedPath(terragruntOptionsSrc *options.TerragruntOptions) (*options.TerragruntOptions, error) {

	pretty.Print(terragruntOptionsSrc)

	newDirName := terragruntOptionsSrc.MvDestination
	oldWorkingDir := terragruntOptionsSrc.WorkingDir
	newWorkingDir, err := filepath.Abs(filepath.Join(filepath.Dir(oldWorkingDir), newDirName))
	if err != nil {
		return nil, err
	}

	newPath := strings.Replace(terragruntOptionsSrc.TerragruntConfigPath, oldWorkingDir, newWorkingDir, 1)

	terragruntOptionsTarget := terragruntOptionsSrc.Clone(newPath) // we leave the new path empty and override it later ourselves, since the Clone function sets the working dir to the parent

	pretty.Print(terragruntOptionsTarget)

	//terragruntOptionsTarget.WorkingDir = newWorkingDir
	//terragruntOptionsTarget.TerragruntConfigPath = strings.Replace(terragruntOptionsSrc.TerragruntConfigPath, oldWorkingDir, newWorkingDir, 1)
	//terragruntOptionsTarget.DownloadDir = strings.Replace(terragruntOptionsSrc.DownloadDir, oldWorkingDir, newWorkingDir, 1)

	return terragruntOptionsTarget, nil
}

func cleanTmpFiles(stack *Stack, filename string) error {
	for _, module := range stack.Modules {
		stateFile := path.Join(module.Path, filename)
		if util.FileExists(stateFile) {
			if err := os.Remove(stateFile); err != nil {
				return errors.WithStackTrace(err)
			}
		}
	}
	return nil
}
