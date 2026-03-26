package processor

import "fmt"

const (
	logIDPrefixLen = 8
	logIDSuffixLen = 4
)

// summarizeIdentifier shortens potentially long identifiers in logs.
//
// This keeps logs operator-friendly for high-throughput paths while still
// preserving enough entropy to correlate related records during debugging.
func summarizeIdentifier(value string) string {
	if value == "" {
		return "<empty>"
	}

	if len(value) <= logIDPrefixLen+logIDSuffixLen+3 {
		return value
	}

	prefix := value[:logIDPrefixLen]
	suffix := value[len(value)-logIDSuffixLen:]
	return fmt.Sprintf("%s...%s", prefix, suffix)
}
