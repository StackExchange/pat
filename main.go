package main

// A wrapper for "puppet agent -t" that enforces rules about disable messages and so on.

import (
	"fmt"
	"os"

	"github.com/StackExchange/pat/version"
	"github.com/urfave/cli"
)

// preprocessArgs inserts a "--disable-message" flag in front of a bare message.
func preprocessArgs(items []string) []string {
	// NOTE(mhenderson): We have an issue with Go flags and the --disable flag. You can specify it like a bool, OR like a string.
	// Because that doesn't play nicely, we define is solely as a bool in the program, but if we find a string value
	// after it in the os.Args() then we insert --disable-message before the message,
	// so that the message itself falls into
	// a new flag. Gross, I know, but it works.
	// TODO(tlim): An cleaner way to do this is to process the flags
	// then let all the remaining arguments be the message.  Remaining arguments
	// are like the "one", "two", "three" in : cat -v -x one two three
	//  If --disable-message != "" and len(remaining_args) != 0; error!
	//  If --disable-message == "", then the message is strings.Join(remaining_args, " ")
	//  Now if the message is still "", substitute the default message.
	// If -disable=true, error out if it wasn't the last flag on the command line.
	// How to get the remaining arguments requires defining a function that
	// gathers them.  See https://github.com/urfave/cli#arguments
	var newArgs []string
	var disabledPosition int
	for i, v := range items {
		if v == "--disable" {
			disabledPosition = i
		}
		if i == disabledPosition+1 && string(v[0]) != "-" {
			newArgs = append(newArgs, "--disable-message", fmt.Sprintf("\"%s\"", v))
			continue
		}
		newArgs = append(newArgs, v)
	}
	return newArgs
}

// main() sets up the application and executes it. The rest of the work is in pat.go
func main() {
	pat := cli.NewApp()
	pat.Name = "pat"
	pat.Usage = "A wrapper for \"puppet agent -t\" (hence the name: P... A... T) that enforces rules about disable messages and so on"
	pat.Version = version.GetVersionInfo()
	pat.UsageText = fmt.Sprintf("%s [flags] [additional puppet commands]", os.Args[0])
	pat.Authors = []cli.Author{
		cli.Author{
			Name:  "Tom Limoncelli",
			Email: "tlimoncelli@stackoverflow.com",
		},
		cli.Author{
			Name:  "Mark Henderson",
			Email: "mhenderson@stackoverflow.com",
		},
	}
	pat.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "disable",
			Usage: "Disable puppet runs",
		},
		cli.StringFlag{
			Name:   "disable-message",
			Value:  "",
			Usage:  "The reason why puppet is being disabled",
			Hidden: true,
		},

		cli.BoolFlag{
			Name:  "enable",
			Usage: "Enable puppet runs",
		},
		cli.BoolFlag{
			Name:  "once",
			Usage: "Run puppet. If puppet was disabled, re-disable when done",
		},
		cli.BoolFlag{
			Name:  "status",
			Usage: "Report disable status",
		},
		cli.BoolFlag{
			Name:  "noop, n",
			Usage: "Pass --noop flag to puppet",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Pass --debug flag to puppet",
		},
		cli.BoolFlag{
			Name:  "timestamp, ts",
			Usage: "Typically used with --debug. Outputs timestamps on all messages",
		},
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Pass --verbose flag to puppet",
		},
		cli.StringFlag{
			Name:  "server",
			Usage: "Pass --server [value] flag to puppet",
		},
		cli.StringFlag{
			Name:  "environment, env, e",
			Usage: "Pass --environment flag to puppet",
		},
		cli.BoolFlag{
			Name:  "facts",
			Usage: "Run puppet facts instead of puppet agent",
		},
	}
	cli.AppHelpTemplate = fmt.Sprintf(`%s
		
SAMPLE USAGE:
	pat
		Runs 'puppet agent -t'
	pat --noop
		Runs 'puppet agent -t --noop'
	pat -e envname
	pat --env envname
		Runs 'puppet agent -t --environment envname' 

	pat --disable message
	pat --disable
	pat -s 3h --disable
		Runs 'puppet --disable' with message, or will prompt for one
		if left blank.
		Silences puppet.left.disabled for 1h or the value set by -s.

	pat --enable
		Runs 'puppet --enable'

	pat --once
		Runs 'puppet agent -t' once.  If Puppet is disabled, it first enables
		it and the re-disables it (whether puppet ran successfully or not).
		Retains the old disable message.
		Silences puppet.left.disabled for 1h or the value set by -s.

	pat --status
		Reveals whether Puppet is enabled/disabled.

NOTES:
	* %s
	* If you want to add regular "puppet agent" flags, add them after '--'.
`, cli.AppHelpTemplate, osRootMessage)

	pat.Action = doPat
	err := pat.Run(preprocessArgs(os.Args))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
