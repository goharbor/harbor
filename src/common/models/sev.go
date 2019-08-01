package models

// Severity represents the severity of a image/component in terms of vulnerability.
type Severity int64

// Sevxxx is the list of severity of image after scanning.
const (
	_ Severity = iota
	SevNone
	SevUnknown
	SevLow
	SevMedium
	SevHigh
)

// String is the output function for severity variable
func (sev Severity) String() string {
	name := []string{"negligible", "unknown", "low", "medium", "high"}
	i := int64(sev)
	switch {
	case i >= 1 && i <= int64(SevHigh):
		return name[i-1]
	default:
		return "unknown"
	}
}
