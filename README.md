# pat (golang) ![finger puppets](docs/puppets-48.png)

Puppet Agent Test tool.

This tool is a wrapper around `puppet agent --test` commands. It simplifies usage
allowing you to do things such as:

- Enable puppet, do a run, and disable puppet all in one step
- When disabling puppet, silence Bosun for an appropriate period of time
- Easily apply different environments and noop's
- Easily test puppet runs against a different master

This is the Go re-implementation of the bash script by the same name, used internally at Stack Overflow.
This version works on both Linux and Windows.

On Windows, this tool is the only way to execute `pat`. It is installed as `C:\Program Files\stack\pat.exe`, and is included in the Windows path variable.

The pat icon is designed by [entertainment from Flaticon](https://www.shareicon.net/show-curtains-entertainment-stage-puppet-puppets-puppet-show-822813)

----

```
NAME:
   pat - A wrapper for "puppet agent -t" (hence the name: P... A... T) that enforces rules about setting silences, disable messages, and so on

USAGE:
   C:\...\...\bin\pat.exe [flags] [additional puppet commands]

VERSION:
   x.y.z (abcdefg) built 2018-01-16T14:01:47Z

AUTHORS:
   Tom Limoncelli <tlimoncelli@stackoverflow.com>
   Mark Henderson <mhenderson@stackoverflow.com>

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --disable                                   Disable puppet runs
   --enable                                    Enable puppet runs
   --once                                      Run puppet. If puppet was disabled, re-disable when done
   --nosilence                                 Do not set a silence when disabling puppet
   --status                                    Report disable status
   --noop, -n                                  Pass --noop flag to puppet
   --debug                                     Pass --debug flag to puppet
   --timestamp, --ts                           Typically used with --debug. Outputs timestamps on all messages
   --verbose                                   Pass --verbose flag to puppet
   --server value                              Pass --server [value] flag to puppet
   --environment value, --env value, -e value  Pass --environment flag to puppet
   --facts                                     Run puppet facts instead of puppet agent
   -s value                                    Set the silence duration to [value] (default: "1h")
   --help, -h                                  show help
   --version, -v                               print the version

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

  pat --nosilence --disable
    Runs 'puppet --disable' but does not silence bosun.

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
  * If not run as administrator, the run will fail immediately.
  * If you want to add regular "puppet agent" flags, add them after '--'.
  * No silence it set if --noop set.```
