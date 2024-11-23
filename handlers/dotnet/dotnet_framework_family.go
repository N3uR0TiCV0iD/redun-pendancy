package dotnet

import (
	"fmt"
	"redun-pendancy/utils"
	"regexp"
	"strings"
)

type DotNetFrameworkFamily struct {
	Regex       *regexp.Regexp
	VersionFunc func(string) string
}

var frameworkFamilies = []DotNetFrameworkFamily{
	{
		//.NET Core
		Regex:       regexp.MustCompile(`(?i)^(net\d+\.\d+|\.netcoreapp.+)$`), // Matches: netX.Y && .NETCoreApp*,
		VersionFunc: extractNetCoreVersion,
	},
	{
		//.NET Framework
		Regex:       regexp.MustCompile(`(?i)^(\.netframework.+|net\d\d)$`), // Matches: .NETFramework* and netXY
		VersionFunc: extractNetFrameworkVersion,
	},
	{
		//.NET Standard
		Regex:       regexp.MustCompile(`(?i)^\.netstandard.+$`), // Matches: .NETStandard*
		VersionFunc: extractNetStandardVersion,
	},
}

func (family *DotNetFrameworkFamily) ContainsFramework(framework string) bool {
	return family.Regex.MatchString(framework)
}

func extractNetCoreVersion(framework string) string {
	framework = strings.ToLower(framework)
	version, hadPrefix := utils.TrimPrefixIfMatch(framework, ".netcoreapp")
	if hadPrefix {
		return version
	}
	version, hadPrefix = utils.TrimPrefixIfMatch(framework, "net")
	if hadPrefix {
		return version
	}
	panic("Failed to extract .NET Core framework version")
}

func extractNetFrameworkVersion(framework string) string {
	framework = strings.ToLower(framework)
	version, hadPrefix := utils.TrimPrefixIfMatch(framework, ".netframework")
	if hadPrefix {
		return version
	}
	version, hadPrefix = utils.TrimPrefixIfMatch(framework, "net")
	if hadPrefix {
		major := version[0:1]
		minor := version[1:2]
		return fmt.Sprintf("%s.%s", major, minor)
	}
	panic("Failed to extract .NET Framework version")
}

func extractNetStandardVersion(framework string) string {
	framework = strings.ToLower(framework)
	version, hadPrefix := utils.TrimPrefixIfMatch(framework, ".netstandard")
	if hadPrefix {
		return version
	}
	panic("Failed to extract .NET Standard framework version")
}
