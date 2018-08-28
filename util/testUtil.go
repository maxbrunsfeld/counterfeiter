package util

import (
	"github.com/xlab/handysort"
	"runtime"
	"sort"
)

func IsLaterThanVersion(version string) bool {
	versions := []string{runtime.Version(), version}
	sort.Sort(handysort.Strings(versions))
	return versions[0] == version
}

