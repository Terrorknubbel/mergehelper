package gitrunner

import (
	"errors"
	"os/exec"
	"log"
	"strings"
	"os"
)

type CommandCheck struct {
	Command      string
	Args         []string
	Output       string
	Expectation	 string
	Forbidden    []string
	ErrorMessage string
}

func CheckPrerequisites(idx int) (bool, string, error) {
	checks := []CommandCheck{
		{Command: "git", Args: []string{"--version"}, Output: "", ErrorMessage: "Git ist nicht installiert."}, // Abtesten
		{Command: "git", Args: []string{"rev-parse", "--is-inside-work-tree"}, Output: "", ErrorMessage: "Du befindest in keinem Git Verzeichnis."},
	}

	check := checks[idx]

	_, err := exec.Command(check.Command, check.Args...).Output()

	hasMore := idx < (len(checks) - 1)

	if err != nil {
		return hasMore, "", errors.New(check.ErrorMessage)
	}

	return hasMore, check.Output, nil
}

func CheckBranchCondition(idx int) (bool, string, error) {
	logFile, _ := os.OpenFile("my_app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer logFile.Close()
	log.SetOutput(logFile)

	checks := []CommandCheck{
		{Command: "git", Args: []string{"branch", "--show-current"}, Output: "Überprüfe Branch…", Expectation: "", Forbidden: []string{"master", "main", "staging"}, ErrorMessage: "Du befindest dich in keinem Feature Branch."},
		{Command: "git", Args: []string{"status", "--porcelain"}, Output: "Prüfe auf Änderungen, die nicht zum Commit vorgesehen sind…", Expectation: "", Forbidden: []string{}, ErrorMessage: "Es gibt Änderungen, die nicht zum Commit vorgesehen sind. Bitte Committe oder Stashe diese vor einem Merge."},
	}

	check := checks[idx]

	output, err := exec.Command(check.Command, check.Args...).Output()
	outputString  := strings.TrimSpace(string(output[:]))
	hasMore := idx < (len(checks) - 1)

	if err != nil {
		return hasMore, "", errors.New(check.ErrorMessage)
	}

	if check.Expectation != outputString {
		return hasMore, "", errors.New(check.ErrorMessage)
	}

	for _, v := range check.Forbidden {
		if outputString == v {
			return hasMore, "", errors.New(check.ErrorMessage)
		}
	}

	return hasMore, check.Output, nil
}
