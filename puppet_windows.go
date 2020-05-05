// +build windows
//go:generate goversioninfo -icon=build/pat.ico -manifest build/pat.exe.manifest

package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	osRootName       = "administrator"
	osRootMessage    = "If not run as administrator, the run will fail immediately."
	osPuppetLockFile = "C:/ProgramData/PuppetLabs/puppet/cache/state/agent_disabled.lock"
	osPuppetBinPath  = "C:/Program Files/Puppet Labs/Puppet/bin/puppet.bat"
)

func isRoot() bool {
	//This is enforced by resource.syso, which embeds pat.exe.manifest during the build process executed
	// by build/build.go. These two things together force Windows to tell it to
	// elevate itself to administrator when it executes.
	return true
}

// There are multiple bug reports with Golang about executing a command with spaces in the name,
// when you also have quoted parameters. To work around this stupid, stupid, stupid problem we can
// write a batch file that executes our command instead, and execute the batch file. Stupid.
// TODO(tlim): This may be fixed in Go 1.10.
func osMakeExec(osPuppetBinPath string, puppetArgs ...string) (*exec.Cmd, error) {
	runBatch := tempFileName("pat-", ".bat")

	batchContents := fmt.Sprintf(`@"%s" %s`, osPuppetBinPath, strings.Join(puppetArgs, " "))
	err := ioutil.WriteFile(runBatch, []byte(batchContents), 0700)
	if err != nil {
		return nil, err
	}
	if isDebug {
		tsLn("DEBUG: Batch file contents:", batchContents)
	}

	cmd := exec.Command(runBatch)
	return cmd, nil
}

// We don't want to leave hundreds of batch files lying around, so clean them up
func osCleanupExec(cmd *exec.Cmd) {
	if isDebug {
		tsLn("DEBUG: Removing", cmd.Path)
	}
	os.Remove(cmd.Path)
}

// https://stackoverflow.com/questions/28005865/golang-generate-unique-filename-with-extension
func tempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
}
