package gitrunner

import (
	"testing"
	"os"
)

// TODO: Testing Framework?

func Setup() (string, string){
	tempDir, _ := os.MkdirTemp("", "")

	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)

	return tempDir, originalDir
}

func Teardown(tempDir string, originalDir string) {
	os.RemoveAll(tempDir)
	os.Chdir(originalDir)
}

func TestCheckPrerequisites(t *testing.T) {
	t.Run("Not in a Git directory", func(t *testing.T) {
		var tempDir, originalDir string = Setup()

		_, _, err := CheckPrerequisites(1)

		if err.Error() != "Du befindest in keinem Git Verzeichnis." {
			t.Errorf(err.Error())
		}

		Teardown(tempDir, originalDir)
	})
}
