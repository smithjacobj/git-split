package git

import (
	"bytes"
	"os/exec"
)

// ShowRef gets the description and patch for the specified commit ref. If it succeeds, buf contains
// the patch text and err is nil. If it fails, buf contains the error output and err contains the
// error returned from cmd.Run().
func ShowRef(ref string) (buf *bytes.Buffer, err error) {
	buf = &bytes.Buffer{}
	cmd := exec.Command("git", "show", ref, "-p", "--no-color")
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
