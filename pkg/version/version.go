package version

import (
	"fmt"
	"runtime/debug"

	goversion "github.com/hashicorp/go-version"
)

var (
	BuildTime  = ""
	version, _ = goversion.NewVersion("0.0.1")
)

func GreatThan(v1, v2 string) bool {
	v1v, _ := goversion.NewVersion(v1)
	v2v, _ := goversion.NewVersion(v2)

	return v1v.GreaterThan(v2v)
}

func GetVersion() string {
	return version.String()
}

func GetBuildTime() string {
	return BuildTime
}

func GetCommitHash() string {
	info, _ := debug.ReadBuildInfo()
	var rev string = "<none>"
	var dirty string = ""
	for _, v := range info.Settings {
		if v.Key == "vcs.revision" {
			rev = v.Value
		}
		if v.Key == "vcs.modified" {
			if v.Value == "true" {
				dirty = "-dirty"
			} else {
				dirty = ""
			}
		}
	}
	return rev + dirty
}

func Print() {
	fmt.Printf("version:%s, build time:%s\nhashid:%s\n", GetVersion(), GetBuildTime(), GetCommitHash())
}
