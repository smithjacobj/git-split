package git

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// ShowRefDescription gets the description for the specified commit ref. If it succeeds, s contains
// the description and err is nil. If it fails, s contains the error output and err contains the
// error returned from Run().
func FormatShowRefDescription(ref, format string) (s string, err error) {
	buf := &bytes.Buffer{}
	cmd := exec.Command("git", "show", ref, "--no-patch", "--no-color", fmt.Sprintf("--format=%s", format))
	cmd.Stdout = buf
	cmd.Stderr = buf

	err = cmd.Run()
	return buf.String(), err
}

// Diff shows the diff/patch between two specific commits. If it succeeds, buf contains the patch
// and err is nil. If it fails, buf contains the error output and err contains the error returned
// from Run()
func Diff(ref1, ref2 string) (buf *bytes.Buffer, err error) {
	buf = &bytes.Buffer{}
	cmd := exec.Command("git", "diff", ref1, ref2, "-p", "--no-color")
	cmd.Stdout = buf
	cmd.Stderr = buf

	err = cmd.Run()
	return
}

func IsDifferent(ref1, ref2 string) (bool, error) {
	buf, err := Diff(ref1, ref2)
	if err != nil {
		return true, err
	} else if buf.Len() == 0 {
		return false, nil
	}
	return true, nil
}

// ApplyPatch applies the patch in buf to the working tree but doesn't add or commit it.
func ApplyPatch(r io.Reader) error {
	// we use --recount instead of trying to manually fix patch chunks ourselves
	cmd := exec.Command("git", "apply", "--recount", "-")
	cmd.Stdin = r

	output, err := cmd.CombinedOutput()
	if err != nil {
		asExecuted := cmd.String()
		return fmt.Errorf("%s: %s\n%s", err, asExecuted, output)
	}
	return nil
}

// HasChanges returns true if there are changes that have not been committed in the working tree
func HasChanges() (bool, error) {
	buf := &bytes.Buffer{}
	cmd := exec.Command("git", "status", "-s")
	cmd.Stdout = buf

	err := cmd.Run()
	if err != nil {
		return true, fmt.Errorf("error running `git status -s`: %s", err)
	}

	var line string
	for reader := bufio.NewReader(buf); err == nil; line, err = reader.ReadString('\n') {
		if len(line) == 0 {
			continue
		}
		line = strings.TrimSpace(line)
		switch line[0] {
		case '?':
			continue
		default:
			return true, nil
		}
	}
	return false, nil
}

// GetCurrentBranchName gets the current branch name
func GetCurrentBranchName() (name string, err error) {
	if output, err := exec.Command("git", "branch", "--show-current").CombinedOutput(); err != nil {
		return "", fmt.Errorf("%s: %s", err, output)
	} else {
		return strings.TrimSpace(string(output)), nil
	}
}

// Commit triggers a commit, bringing up the default editor with the specified message
func Commit(message string) error {
	cmd := exec.Command("git", "commit", "-F", "-")
	cmd.Stdin = strings.NewReader(message)

	if output, err := cmd.CombinedOutput(); err != nil {
		asExecuted := cmd.String()
		return fmt.Errorf("%s: %s\n%s", err, asExecuted, output)
	}
	return nil
}

// Amend runs `git commit --amend` to amend the details of the last commit
func Amend() error {
	cmd := exec.Command("git", "commit", "--amend")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Checkout the specified ref
func Checkout(ref string) error {
	cmd := exec.Command("git", "checkout", ref)
	if output, err := cmd.CombinedOutput(); err != nil {
		asExecuted := cmd.String()
		return fmt.Errorf("%s: %s\n%s", err, asExecuted, output)
	}
	return nil
}

// CreateAndSwitchToBranch creates a new branch and switches to it (`git checkout -b`)
func CreateAndSwitchToBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	if output, err := cmd.CombinedOutput(); err != nil {
		asExecuted := cmd.String()
		return fmt.Errorf("%s: %s\n%s", err, asExecuted, output)
	}
	return nil
}

// RevParse gets the hash for a ref
func RevParse(ref string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", ref)
	if output, err := cmd.CombinedOutput(); err != nil {
		asExecuted := cmd.String()
		return "", fmt.Errorf("%s: %s\n%s", err, asExecuted, output)
	} else {
		return string(output), nil
	}
}

// Add does a `git add`
func Add(paths ...string) error {
	arg := append([]string{"add", "--"}, paths...)
	cmd := exec.Command("git", arg...)

	if output, err := cmd.CombinedOutput(); err != nil {
		asExecuted := cmd.String()
		return fmt.Errorf("%s: %s\n%s", err, asExecuted, output)
	}
	return nil
}
