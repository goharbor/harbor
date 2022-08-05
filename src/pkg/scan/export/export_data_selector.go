package export

import "github.com/bmatcuk/doublestar"

const (
	CVEIDMatches      = "cveIdMatches"
	PackageMatches    = "packageMatches"
	ScannerMatches    = "scannerMatches"
	CVE2VectorMatches = "cve2VectorMatches"
	CVE3VectorMatches = "cve3VectorMatches"
)

// VulnerabilityDataSelector is a specialized implementation of a selector
// leveraging the doublestar pattern to select vulnerabilities
type VulnerabilityDataSelector interface {
	Select(vulnDataRecords []Data, decoration string, pattern string) ([]Data, error)
}

type defaultVulnerabilitySelector struct{}

// NewVulnerabilityDataSelector selects the vulnerability data record
// that matches the provided conditions
func NewVulnerabilityDataSelector() VulnerabilityDataSelector {
	return &defaultVulnerabilitySelector{}
}

func (vds *defaultVulnerabilitySelector) Select(vulnDataRecords []Data, decoration string, pattern string) ([]Data, error) {
	selected := make([]Data, 0)
	value := ""

	for _, vulnDataRecord := range vulnDataRecords {
		switch decoration {
		case CVEIDMatches:
			value = vulnDataRecord.CVEId
		case PackageMatches:
			value = vulnDataRecord.Package
		case ScannerMatches:
			value = vulnDataRecord.ScannerName
		}
		matched, err := vds.match(pattern, value)
		if err != nil {
			return nil, err
		}
		if matched {
			selected = append(selected, vulnDataRecord)
		}
	}
	return selected, nil
}

func (vds *defaultVulnerabilitySelector) match(pattern, str string) (bool, error) {
	if len(pattern) == 0 {
		return true, nil
	}
	return doublestar.Match(pattern, str)
}
