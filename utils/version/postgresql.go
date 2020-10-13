package version

import (
	"regexp"
	"strings"
)

// regexps to extract version numbers from the `SELECT version()` output
var (
	postgresDBRegexp = regexp.MustCompile(`PostgreSQL ([\d\.]+)`)
)

func ParsePostgreSQLVersion(v string) string {
	m := postgresDBRegexp.FindStringSubmatch(v)
	if len(m) != 2 {
		return ""
	}

	parts := strings.Split(m[1], ".")
	switch len(parts) {
	case 1: // major only
		return parts[0]
	case 2: // major and patch
		return parts[0]
	case 3: // major, minor, and patch
		return parts[0] + "." + parts[1]
	default:
		return ""
	}
}
