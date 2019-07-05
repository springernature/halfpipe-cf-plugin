package helpers

import (
	"strings"
	"fmt"
)

func CreateCandidateAppName(appName string) string {
	return strings.Join([]string{appName, "CANDIDATE"}, "-")
}

func CreateCandidateHostname(appName string, space string) string {
	return strings.Join([]string{appName, space, "CANDIDATE"}, "-")
}

func CreateOldAppName(appName string) string {
	return strings.Join([]string{appName, "OLD"}, "-")
}

func CreateDeleteName(appName string, index int) string {
	if index == 0 {
		return fmt.Sprintf("%s-DELETE", appName)
	}
	return fmt.Sprintf("%s-DELETE-%d", appName, index)
}
