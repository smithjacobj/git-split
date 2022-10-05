package git

import (
	"bytes"
	"os/exec"
)

// GetPatchForRef gets the patch format diff between ref and the previous commit. If it succeeds,
// buf contains the patch text and err is nil. If it fails, buf contains the error output and err
// contains the error returned from cmd.Run().
func GetPatchForRef(ref string) (buf *bytes.Buffer, err error) {
	buf = &bytes.Buffer{}
	cmd := exec.Command("git", "show", ref, "-p", "--format=''")
	cmd.Stdout = buf
	cmd.Stderr = buf

	err = cmd.Run()
	return
}

// GetLogForRef gets the log message (subject and body) between ref and the previous commit. If it
// succeeds, buf contains the log message and err is nil. If it fails, buf contains the error output
// and err contains the error returned from cmd.Run().
func GetLogForRef(ref string) (buf *bytes.Buffer, err error) {
	buf = &bytes.Buffer{}
	cmd := exec.Command("git", "show", ref, "-s", "--format=%s\n%B")
	cmd.Stdout = buf
	cmd.Stderr = buf

	err = cmd.Run()
	return
}

// ApplyPatch applies the patch in buf to the working tree but doesn't add or commit it.
func ApplyPatch(buf *bytes.Buffer) error {
	cmd := exec.Command("git", "apply", "-")
	cmd.Stdin = buf
	return nil
}
