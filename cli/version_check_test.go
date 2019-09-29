package cli

import (
	"github.com/gruntwork-io/terragrunt/errors"
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckTerraformVersionMeetsConstraintEqual(t *testing.T) {
	t.Parallel()
	testCheckTerraformVersionMeetsConstraint(t, "v0.9.3", ">= v0.9.3", true)
}

func TestCheckTerraformVersionMeetsConstraintGreaterPatch(t *testing.T) {
	t.Parallel()
	testCheckTerraformVersionMeetsConstraint(t, "v0.9.4", ">= v0.9.3", true)
}

func TestCheckTerraformVersionMeetsConstraintGreaterMajor(t *testing.T) {
	t.Parallel()
	testCheckTerraformVersionMeetsConstraint(t, "v1.0.0", ">= v0.9.3", true)
}

func TestCheckTerraformVersionMeetsConstraintLessPatch(t *testing.T) {
	t.Parallel()
	testCheckTerraformVersionMeetsConstraint(t, "v0.9.2", ">= v0.9.3", false)
}

func TestCheckTerraformVersionMeetsConstraintLessMajor(t *testing.T) {
	t.Parallel()
	testCheckTerraformVersionMeetsConstraint(t, "v0.8.8", ">= v0.9.3", false)
}

func TestParseTerraformVersionNormal(t *testing.T) {
	t.Parallel()
	testParseTerraformVersion(t, "Terraform v0.9.3", "v0.9.3", nil)
}

func TestParseTerraformVersionWithoutV(t *testing.T) {
	t.Parallel()
	testParseTerraformVersion(t, "Terraform 0.9.3", "0.9.3", nil)
}

func TestParseTerraformVersionWithDebug(t *testing.T) {
	t.Parallel()
	testParseTerraformVersion(t, "Terraform v0.9.4 cad024a5fe131a546936674ef85445215bbc4226", "v0.9.4", nil)
}

func TestParseTerraformVersionWithChanges(t *testing.T) {
	t.Parallel()
	testParseTerraformVersion(t, "Terraform v0.9.4-dev (cad024a5fe131a546936674ef85445215bbc4226+CHANGES)", "v0.9.4", nil)
}

func TestParseTerraformVersionWithDev(t *testing.T) {
	t.Parallel()
	testParseTerraformVersion(t, "Terraform v0.9.4-dev", "v0.9.4", nil)
}

func TestParseTerraformVersionInvalidSyntax(t *testing.T) {
	t.Parallel()
	testParseTerraformVersion(t, "invalid-syntax", "", InvalidTerraformVersionSyntax("invalid-syntax"))
}

func testCheckTerraformVersionMeetsConstraint(t *testing.T, currentVersion string, versionConstraint string, versionMeetsConstraint bool) {
	current, err := version.NewVersion(currentVersion)
	if err != nil {
		t.Fatalf("Invalid current version specified in test: %v", err)
	}

	err = checkTerraformVersionMeetsConstraint(current, versionConstraint)
	if versionMeetsConstraint && err != nil {
		assert.Nil(t, err, "Expected Terraform version %s to meet constraint %s, but got error: %v", currentVersion, versionConstraint, err)
	} else if !versionMeetsConstraint && err == nil {
		assert.NotNil(t, err, "Expected Terraform version %s to NOT meet constraint %s, but got back a nil error", currentVersion, versionConstraint)
	}
}

func testParseTerraformVersion(t *testing.T, versionString string, expectedVersion string, expectedErr error) {
	actualVersion, actualErr := parseTerraformVersion(versionString)
	if expectedErr == nil {
		expected, err := version.NewVersion(expectedVersion)
		if err != nil {
			t.Fatalf("Invalid expected version specified in test: %v", err)
		}

		assert.Nil(t, actualErr)
		assert.Equal(t, expected, actualVersion)
	} else {
		assert.True(t, errors.IsError(actualErr, expectedErr))
	}
}
