package plans

import (
	"strings"
)

func createCandidateAppName(appName string) string {
	return strings.Join([]string{appName, "CANDIDATE"}, "-")
}
