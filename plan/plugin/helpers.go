package plugin

import (
	"strings"
	"fmt"
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

func createDeleteName(appName string, index int) string {
	if index == 0 {
		return fmt.Sprintf("%s-DELETE", appName)
	}
	return fmt.Sprintf("%s-DELETE-%d", appName, index)
}
