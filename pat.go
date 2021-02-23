package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
	"time"

	silence "github.com/StackExchange/pat/addsilence"
	"github.com/urfave/cli"
)

var (
	enableSilence       = true
	additionalArguments = []string{}
	isDebug             = false
	isTimestamp         = false
	isNoop              = false
	isFacts             = false
	puppetDisabled      = false
)

func doPat(pat *cli.Context) error {
	//Throw these as globals as they're used all over the place. Saves us from passing pat through everywhere.
	additionalArguments = pat.Args()
	enableSilence = !pat.Bool("nosilence")

	//We need some additional arguments for dealing with things like the verbose and debug flags
	additionalArguments = pat.Args()
	var flagArguments []string
	if pat.Bool("timestamp") {
		isTimestamp = true
	}
	if pat.Bool("noop") {
		tsLn("NOOP mode enabled")
		isNoop = true
		flagArguments = append(flagArguments, "--noop", "--agent_disabled_lockfile", "/non_existant")
	}
	if pat.Bool("debug") {
		isDebug = true
		flagArguments = append(flagArguments, "--debug")
	}
	if pat.Bool("facts") {
		isFacts = true
	}
	if pat.Bool("verbose") {
		flagArguments = append(flagArguments, "--verbose")
	}
	if pat.Bool("verbose") {
		flagArguments = append(flagArguments, "--verbose")
	}
	if pat.IsSet("environment") {
		flagArguments = append(flagArguments, "--environment", pat.String("environment"))
	}
	if pat.IsSet("server") {
		flagArguments = append(flagArguments, "--server", pat.String("server"))
	}
	additionalArguments = append(flagArguments, additionalArguments...)

	//Check if puppet is currently disabled
	if _, err := os.Stat(osPuppetLockFile); err == nil {
		puppetDisabled = true
	}

	if isDebug {
		tsLn("DEBUG: -- Flags --")
		tsLn("DEBUG: disable:", pat.String("disable"))
		tsLn("DEBUG: enable:", pat.Bool("enable"))
		tsLn("DEBUG: once:", pat.Bool("once"))
		tsLn("DEBUG: nosilence:", pat.Bool("nosilence"))
		tsLn("DEBUG: status:", pat.Bool("status"))
		tsLn("DEBUG: noop:", pat.Bool("noop"))
		tsLn("DEBUG: debug:", pat.Bool("debug"))
		tsLn("DEBUG: timestamp:", pat.Bool("timestamp"))
		tsLn("DEBUG: verbose:", pat.Bool("verbose"))
		tsLn("DEBUG: server:", pat.String("server"))
		tsLn("DEBUG: environment:", pat.String("environment"))
		tsLn("DEBUG: s:", pat.String("s"))

		tsLn("DEBUG: -- OS --")
		tsLn("DEBUG: osRootName:", osRootName)
		tsLn("DEBUG: osRootMessage:", osRootMessage)
		tsLn("DEBUG: osPuppetLockFile:", osPuppetLockFile)
		tsLn("DEBUG: osPuppetBinPath:", osPuppetBinPath)

		tsLn("DEBUG: -- PROGRAM --")
		tsLn("DEBUG: puppetDisabled:", puppetDisabled)
		tsLn("DEBUG: additionalArgs:", additionalArguments)
		tsLn("DEBUG: enableSilence:", enableSilence)
	}

	var err error

	// CMD: --status
	if pat.Bool("status") {
		if puppetDisabled {
			disabledMessage, err := getPuppetDisabledMessage()
			if err != nil {
				return err
			}
			tsLn("STATUS: PUPPET IS DISABLED")
			tsLn("DISABLE MESSAGE: ", disabledMessage)
			return nil
		}
		tsLn("STATUS: PUPPET IS ENABLED")
		return nil
	}

	// CMD: --once
	if pat.Bool("once") {
		var disabledMessage string
		var puppetWasDisabled bool
		//If puppet is disabled; enable it
		if puppetDisabled {
			puppetWasDisabled = true
			disabledMessage, err = getPuppetDisabledMessage()
			if err != nil {
				return err
			}
			err = enablePuppet()
			if err != nil {
				return err
			}
		}

		//Do a default run
		execPuppet()

		//If puppet was disabled; disable it again
		if puppetWasDisabled {
			err = disablePuppet(disabledMessage, pat.String("s"))
			if err != nil {
				return err
			}
		}

		return nil

	}

	// CMD: --disable [message]
	if pat.IsSet("disable") {
		err = disablePuppet(pat.String("disable-message"), pat.String("s"))
		return err
	}

	// CMD: --enable
	if pat.IsSet("enable") {
		err = enablePuppet()
		return err
	}

	//Deeeeefault
	err = execPuppet()
	return err
}

func enablePuppet() error {
	err := execPuppet("--enable")
	if err != nil {
		return err
	}
	puppetDisabled = false
	return nil
}

// Disable puppet. If puppet is already disabled, will return an error
func disablePuppet(message, silenceduration string) error {
	if puppetDisabled {
		return fmt.Errorf("Puppet is already disabled")
	}
	//If no message is specified, we have an interactive prompt to ask the user for the message
	if message == "" {
		puppetDisabledMessageDefault := defaultPuppetDisableMessage()
		//Ask the user for their disable message
		reader := bufio.NewReader(os.Stdin)
		tsLn("No disable message specified.")
		fmt.Printf("Enter message or press ENTER for [%s]: ", puppetDisabledMessageDefault)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "" {
			message = puppetDisabledMessageDefault
		} else {
			message = text
		}
		tsLn()
	}

	//Clean a few things up - like removing linebreaks, and appending quotes around the message
	message = strings.Replace(message, "\n", "", -1)

	if string(message[0]) != "\"" {
		message = fmt.Sprintf("\"%s\"", message)
	}

	err := execPuppet("--disable", message)
	if err != nil {
		return err
	}
	err = silenceBosun(silenceduration, message)
	if err != nil {
		return err
	}

	return nil
}

// Executes puppet with the arguments provided, along with fixed arguments, and mixing in the additional
// arguments that were specified on the command line
func execPuppet(args ...string) error {
	puppetArgs := []string{}
	//If we're doing a fact run, we don't need the agent command
	if isFacts {
		puppetArgs = []string{"facts"}
	}
	// When we are disabling puppet, we need to drop the -t
	if len(args) > 1 && args[0] == "--disable" {
		puppetArgs = []string{"agent"}
	}

	// If don't have any overrides, use these defaults
	if len(puppetArgs) == 0 {
		puppetArgs = []string{"agent", "-t"}
	}

	// These are the additional user-supplied arguments
	for _, v := range additionalArguments {
		puppetArgs = append(puppetArgs, v)
	}

	// And we tack the parameter-supplied arguments onto the very end
	puppetArgs = append(puppetArgs, args...)

	if isDebug {
		tsLn("DEBUG: Invoking", osPuppetBinPath, "with arguments:", puppetArgs)
	}

	//Execute and return
	cmd, err := osMakeExec(osPuppetBinPath, puppetArgs...)
	defer osCleanupExec(cmd)
	if err != nil {
		return err
	}
	cmdReaderStd, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	scannerStd := bufio.NewScanner(cmdReaderStd)
	go func() {
		for scannerStd.Scan() {
			tsPrintf("%s\n", scannerStd.Text())
		}
	}()

	cmdReaderErr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	scannerErr := bufio.NewScanner(cmdReaderErr)
	go func() {
		for scannerErr.Scan() {
			tsPrintf("%s\n", scannerErr.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	return err
}

func silenceBosun(duration, message string) error {
	// If silencing is disabled, or we're doing a noop
	if !enableSilence || isNoop {
		if isDebug {
			tsLn("DEBUG: Skipping silence, as silence is disabled or we're doing a noop")
		}
		return nil
	}

	//Call silence here
	hostName, err := os.Hostname()
	if err != nil {
		return err
	}
	hosts := []string{hostName}
	silence, err := silence.EasySilence("puppet.left.disabled", duration, message, hosts)
	if silence != "" {
		tsLn(silence)
	}
	if err != nil {
		return err
	}
	return nil
}

//Create a form message to use when disabling puppet if no message is specified
func defaultPuppetDisableMessage() string {
	currentUser, err := user.Current() //Get the current user for the form message
	if err != nil {
		return ""
	}
	//Default form message
	return fmt.Sprintf("Disabled by %s. If I forget to re-enable it after 2 hours, anyone may re-enable and any problems this causes are my responsibility.", currentUser.Name)
}

func getPuppetDisabledMessage() (string, error) {
	//Check if puppet is disabled so we can later on re-use the disabled message (or not run at all)
	var puppetDisabledMessage string
	//If we have a more specific puppet message, get it and use that instead of the form message
	if _, err := os.Stat(osPuppetLockFile); err == nil {
		lockMessageBytes, err := ioutil.ReadFile(osPuppetLockFile)
		if err != nil {
			return "", err
		}
		var messageStruct disabledMessage
		err = json.Unmarshal(lockMessageBytes, &messageStruct)
		if err != nil {
			return "", err
		}
		puppetDisabledMessage = messageStruct.DisabledMessage
	}

	return puppetDisabledMessage, nil
}

const tsPadding = 35

func ts() string {
	ts := fmt.Sprintf("%s:", time.Now().Format(time.RFC3339Nano))
	if len(ts) < tsPadding {
		ts = ts + strings.Repeat(" ", tsPadding-len(ts))
	}

	return ts
}

func tsPrintf(format string, a ...interface{}) {
	if isTimestamp {
		fmt.Printf("%s", ts())
	}
	fmt.Printf(format, a...)
}

func tsLn(a ...interface{}) {
	if isTimestamp {
		fmt.Printf("%s", ts())
	}
	fmt.Println(a...)
}
