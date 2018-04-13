package plugin

import (
	"strings"
)

func createCandidateAppName(appName string) string {
	return strings.Join([]string{appName, "CANDIDATE"}, "-")
}

func createCandidateHostname(appName string, space string) string {
	return strings.Join([]string{appName, space, "CANDIDATE"}, "-")
}

func createOldAppName(appName string) string {
	return strings.Join([]string{appName, "OLD"}, "-")
}

func createDeleteName(appName string) string {
	return strings.Join([]string{appName, "DELETE"}, "-")
}
