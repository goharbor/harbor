package regtoken

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type HandlerSuite struct {
	suite.Suite
}

func doPullManifestRequest(projectName, name, tag string, next ...http.HandlerFunc) int {
	repository := fmt.Sprintf("%s/%s", projectName, name)

	url := fmt.Sprintf("/v2/%s/manifests/%s", repository, tag)
	req, _ := http.NewRequest("GET", url, nil)

	token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6IkNWUTc6REM3NTpHVEROOkxTTUs6VUFJTjpIUUVWOlZVSDQ6Q0lRRDpRV01COlM0Qzc6U0c0STpGRUhYIn0.eyJpc3MiOiJoYXJib3ItdG9rZW4taXNzdWVyIiwic3ViIjoicm9ib3QkZGVtbzExIiwiYXVkIjoiaGFyYm9yLXJlZ2lzdHJ5IiwiZXhwIjoxNTcxNzYzOTI2LCJuYmYiOjE1NzE3NjM4NjYsImlhdCI6MTU3MTc2Mzg2NiwianRpIjoiTnRaZWx4Z01KTUU1MXlEMCIsImFjY2VzcyI6W3sidHlwZSI6InJlcG9zaXRvcnkiLCJuYW1lIjoibGlicmFyeS9oZWxsby13b3JsZCIsImFjdGlvbnMiOlsicHVzaCIsIioiLCJwdWxsIiwic2Nhbm5lcnB1bGwiXX1dfQ.GlWuvtoxmChnpvbWaG5901Z9-g63DrzyNUREWlDbR5gnNeuOKjLNyE4QpogAQKx2yYtcGxbqNL3VfJkExJ_gMS0Qw8e10utGOawwqD4oqf_J06eKq4HzpZJengZfcjMA4g2RoeOlqdVdwimB_PdX9vkBO1od0wX0Cc2v0p2w5TkibcThKRoeLeVs2oRewkKLuVHNSM8wwRIlAvpWJuNnvRCFlHRkLcZM_KpGXqT7H-PZETTisWCi1pMxeYEwIsDFLlTKdV8LaiDeDmH-RaLOsuyAySYEW9Ynk5K3P_dUl2c_SYQXloPyi0MvXxSn6EWE4eHF2oQDM_SvIzR9sOVB8TtjMjKKMQ4yr_mqgMcfEpnInJATExBR56wmxNdLESncHl8rUYCe2jCjQFuR9NGQA1tGdjI4NoBN-OVD0dBs9rm_mkb2tgD-3gEhyzAw6hg0uzDsF7bj5Aq8scoi42UurhX2bZM89s4-TWBp4DWuBG0HDiwpOiBvB3RMm6MpQxsqrl0hQm_WH18L6QCknAW2e3d_6DJWJ0eBzISrhDr7LkqJKl1J8pv4zqoh_EUVeLyzTmjEULm-VbnpVF4wW5yTLF3S6F7Ox4vwWtVfi1XQNVOcJDB3VPUsRgiTTuCW-ZGcBLw-OdIcwaJ3T_QZkEjUw1f6i1JcGa0Mpgl83aLiSdQ 0xc0003c77c0 map[alg:RS256 kid:CVQ7:DC75:GTDN:LSMK:UAIN:HQEV:VUH4:CIQD:QWMB:S4C7:SG4I:FEHX typ:JWT] 0xc000496000 GlWuvtoxmChnpvbWaG5901Z9-g63DrzyNUREWlDbR5gnNeuOKjLNyE4QpogAQKx2yYtcGxbqNL3VfJkExJ_gMS0Qw8e10utGOawwqD4oqf_J06eKq4HzpZJengZfcjMA4g2RoeOlqdVdwimB_PdX9vkBO1od0wX0Cc2v0p2w5TkibcThKRoeLeVs2oRewkKLuVHNSM8wwRIlAvpWJuNnvRCFlHRkLcZM_KpGXqT7H-PZETTisWCi1pMxeYEwIsDFLlTKdV8LaiDeDmH-RaLOsuyAySYEW9Ynk5K3P_dUl2c_SYQXloPyi0MvXxSn6EWE4eHF2oQDM_SvIzR9sOVB8TtjMjKKMQ4yr_mqgMcfEpnInJATExBR56wmxNdLESncHl8rUYCe2jCjQFuR9NGQA1tGdjI4NoBN-OVD0dBs9rm_mkb2tgD-3gEhyzAw6hg0uzDsF7bj5Aq8scoi42UurhX2bZM89s4-TWBp4DWuBG0HDiwpOiBvB3RMm6MpQxsqrl0hQm_WH18L6QCknAW2e3d_6DJWJ0eBzISrhDr7LkqJKl1J8pv4zqoh_EUVeLyzTmjEULm-VbnpVF4wW5yTLF3S6F7Ox4vwWtVfi1XQNVOcJDB3VPUsRgiTTuCW-ZGcBLw-OdIcwaJ3T_QZkEjUw1f6i1JcGa0Mpgl83aLiSdQ"
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	rr := httptest.NewRecorder()

	af := &middleware.ArtifactInfo{
		Repository: name,
		Reference:  tag,
		Tag:        tag,
		Digest:     "",
	}

	var n http.HandlerFunc
	if len(next) > 0 {
		n = next[0]
	} else {
		n = func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}
	}
	ctx := context.WithValue(req.Context(), middleware.ArtifactInfoKey, af)
	*req = *(req.WithContext(ctx))
	n.ServeHTTP(util.NewCustomResponseWriter(rr), req)

	return rr.Code
}

func (suite *HandlerSuite) TestPullManifest() {
	code1 := doPullManifestRequest("library", "photon", "release-1.10")
	suite.Equal(http.StatusNotFound, code1)
}

func TestMain(m *testing.M) {
	if result := m.Run(); result != 0 {
		os.Exit(result)
	}
}

func TestRunHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}
