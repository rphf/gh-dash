package gitspice

import "os/exec"

// HasSpiceData tells if the git repo at dir is tracked by git-spice, which
// keeps its state in the refs/spice/data ref.
func HasSpiceData(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", "--quiet", "refs/spice/data")
	cmd.Dir = dir
	return cmd.Run() == nil
}
