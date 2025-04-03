package cloud

import (
	"strings"
)

// parseCloudResource splits "type/name" into (type, name)
func parseCloudResource(resource string) (string, string) {
	parts := strings.Split(resource, "/")
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return resource, "" // If no slash, assume it's just the resource type
}
