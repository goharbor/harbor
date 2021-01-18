package exporter

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/stretchr/testify/suite"
)

type ArtifactCollectorTestSuite struct {
	suite.Suite
}

func (t *ArtifactCollectorTestSuite) SetupTest() {
	var (
		testPro1 = models.Project{OwnerID: 1, Name: "test1", Metadata: map[string]string{"public": "true"}}
		testPro2 = models.Project{OwnerID: 1, Name: "test2", Metadata: map[string]string{"public": "false"}}
		// rs1      = qtypes.ResourceList{qtypes.ResourceStorage: 100}
		// rs2      = qtypes.ResourceList{qtypes.ResourceStorage: 200}
		repo1 = models.RepoRecord{Name: "repo1"}
		repo2 = models.RepoRecord{Name: "repo2"}
		pmIDs = []int{}
		art1  = artifact.Artifact{
			RepositoryName: repo1.Name,
			Type:           "IMAGE",
			Digest:         "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180",
			Size:           100,
		}
		art2 = artifact.Artifact{
			RepositoryName: repo1.Name,
			Type:           "IMAGE",
			Digest:         "sha256:3198b18471892718923712837192831287312893712893712897312db1a3bc73",
			Size:           101,
		}

		vuln1 = scan.VulnerabilityRecord{
			CVEID:          "CVE-001",
			Severity:       "High",
			Fix:            "v0.0.2",
			PackageVersion: "v0.0.1",
			Package:        "vuln-package1",
		}
		vuln2 = scan.VulnerabilityRecord{
			CVEID:          "CVE-002",
			Severity:       "Medium",
			Fix:            "v0.0.2",
			PackageVersion: "v0.0.1",
			Package:        "vuln-package2",
		}
		vuln3 = scan.VulnerabilityRecord{
			CVEID:          "CVE-003",
			Severity:       "Low",
			Fix:            "",
			PackageVersion: "v0.0.1",
			Package:        "vuln-package3",
		}
	)

	test.InitDatabaseFromEnv()

}
