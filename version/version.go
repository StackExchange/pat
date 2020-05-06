// Package version holds some version data common to patcher client and server.
// Most of these values will be inserted at build time with `-ldFlags` directives for official builds.
package version // import "github.com/StackExchange/pat/version"

import (
	"fmt"
	"time"
)

// These variables will be set at linking time for official builds.
// build.go will set date and sha, but `go get` will set none of these.

type BuildVersionInfo struct {
	// Version number for official releases Updated manually before each release.
	Version string
	// Set to any non-empty value by official release script
	OfficialBuild string
	//Set the branch name that we are building here
	BuildBranch string
	// VersionDate Date and time of build. Should be in YYYYMMDDHHMMSS format
	VersionDate string
	// VersionSHA should be set at build time as the most recent commit hash.
	VersionSHA string
}

// GetVersionDate returns a go time var from the time string
func (b BuildVersionInfo) GetVersionDate() time.Time {
	thisTime, err := time.Parse("20060102150405", b.VersionDate)
	if err != nil {
		return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	return thisTime
}

var (
	OfficialBuild string
	BuildBranch   string
	VersionDate   string
	VersionSHA    string
	VersionNumber string

	// BuildVersion holds the build details
	BuildVersion = BuildVersionInfo{
		Version:       VersionNumber,
		OfficialBuild: OfficialBuild,
		BuildBranch:   BuildBranch,
		VersionDate:   VersionDate,
		VersionSHA:    VersionSHA,
	}
)

// GetVersionInfo returns a string representing the version information for the current binary.
func GetVersionInfo() string {
	var sha, build string
	version := ShortVersion()
	if buildTime, err := time.Parse("20060102150405", BuildVersion.VersionDate); err == nil {
		build = " built " + buildTime.Format(time.RFC3339)
	}
	if BuildVersion.VersionSHA != "" {
		sha = fmt.Sprintf(" (%s)", BuildVersion.VersionSHA)
	}
	return fmt.Sprintf("%s%s%s", version, sha, build)
}

// ShortVersion returns a short build string, such as "2.0.0-dev"
func ShortVersion() string {
	version := BuildVersion.Version

	if version == "" {
		version = "4.0.0"
	}

	if BuildVersion.OfficialBuild == "" {
		if BuildVersion.BuildBranch != "" {
			version += fmt.Sprintf("-%s", BuildVersion.BuildBranch)
		} else {
			version += "-dev"
		}
	}
	return version
}
