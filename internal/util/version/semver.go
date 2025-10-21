package version

import (
	"fmt"
	"strconv"
	"strings"
)

type SemVer struct{ Major, Minor, Patch int }

func Parse(s string) (SemVer, error) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		return SemVer{}, fmt.Errorf("invalid semver: %s", s)
	}
	maj, _ := strconv.Atoi(parts[0])
	min, _ := strconv.Atoi(parts[1])
	pat := 0
	if len(parts) > 2 {
		pat, _ = strconv.Atoi(parts[2])
	}
	return SemVer{maj, min, pat}, nil
}

func (s SemVer) Compare(o SemVer) int {
	if s.Major != o.Major {
		if s.Major < o.Major {
			return -1
		} else {
			return 1
		}
	}
	if s.Minor != o.Minor {
		if s.Minor < o.Minor {
			return -1
		} else {
			return 1
		}
	}
	if s.Patch != o.Patch {
		if s.Patch < o.Patch {
			return -1
		} else {
			return 1
		}
	}
	return 0
}

func IsBreaking(from, to SemVer) bool { return to.Major > from.Major }
func IsAdditive(from, to SemVer) bool { return to.Major == from.Major && to.Minor > from.Minor }
func IsPatch(from, to SemVer) bool {
	return to.Major == from.Major && to.Minor == from.Minor && to.Patch > from.Patch
}
