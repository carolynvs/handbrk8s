package jobs

import "regexp"

// sanitizeJobName replaces characters that aren't allowed in a k8s name with dashes.
func sanitizeJobName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	return re.ReplaceAllString(name, "-")
}
