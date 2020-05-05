// +build !windows

package main

import (
	"os"
	"os/exec"
)

const (
	osRootName       = "root"
	osRootMessage    = "If not run as root, it will run puppet via sudo automatically."
	osPuppetLockFile = "/opt/puppetlabs/puppet/cache/state/agent_disabled.lock"
	osPuppetBinPath  = "/opt/puppetlabs/bin/puppet"
)

func isRoot() bool {
	return os.Geteuid() == 0
}

// This is the normal way of execing, see the Windows one for the messed up way
func osMakeExec(osPuppetBinPath string, puppetArgs ...string) (*exec.Cmd, error) {
	cmd := exec.Command(osPuppetBinPath, puppetArgs...)
	return cmd, nil
}

// There's nothing to clean up
func osCleanupExec(cmd *exec.Cmd) {
	return
}
