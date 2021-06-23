package authproxy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	cfgModels "github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/usergroup"
	"github.com/goharbor/harbor/src/pkg/usergroup/model"
	k8s_api_v1beta1 "k8s.io/api/authentication/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// TokenReview ...
func TokenReview(rawToken string, authProxyConfig *cfgModels.HTTPAuthProxy) (k8s_api_v1beta1.TokenReviewStatus, error) {

	emptyStatus := k8s_api_v1beta1.TokenReviewStatus{}
	// Init auth client with the auth proxy endpoint.
	authClientCfg := &rest.Config{
		Host: authProxyConfig.TokenReviewEndpoint,
		ContentConfig: rest.ContentConfig{
			GroupVersion:         &schema.GroupVersion{},
			NegotiatedSerializer: serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs},
		},
		TLSClientConfig: getTLSConfig(authProxyConfig),
	}
	authClient, err := rest.RESTClientFor(authClientCfg)
	if err != nil {
		return emptyStatus, err
	}

	// Do auth with the token.
	tokenReviewRequest := &k8s_api_v1beta1.TokenReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TokenReview",
			APIVersion: "authentication.k8s.io/v1beta1",
		},
		Spec: k8s_api_v1beta1.TokenReviewSpec{
			Token: rawToken,
		},
	}
	res := authClient.Post().Body(tokenReviewRequest).Do(context.Background())
	err = res.Error()
	if err != nil {
		log.Errorf("fail to POST auth request, %v", err)
		return emptyStatus, err
	}
	resRaw, err := res.Raw()
	if err != nil {
		log.Errorf("fail to get raw data of token review, %v", err)
		return emptyStatus, err
	}
	// Parse the auth response, check the user name and authenticated status.
	tokenReviewResponse := &k8s_api_v1beta1.TokenReview{}
	err = json.Unmarshal(resRaw, tokenReviewResponse)
	if err != nil {
		log.Errorf("fail to decode token review, %v", err)
		return emptyStatus, err
	}
	return tokenReviewResponse.Status, nil

}

func getTLSConfig(config *cfgModels.HTTPAuthProxy) rest.TLSClientConfig {
	if config.VerifyCert && len(config.ServerCertificate) > 0 {
		return rest.TLSClientConfig{
			CAData: []byte(config.ServerCertificate),
		}
	}
	return rest.TLSClientConfig{
		Insecure: !config.VerifyCert,
	}
}

// UserFromReviewStatus transform a review status to a user model.
// Group entries will be populated if needed.
func UserFromReviewStatus(status k8s_api_v1beta1.TokenReviewStatus, adminGroups []string, adminUsernames []string) (*models.User, error) {
	if !status.Authenticated {
		return nil, fmt.Errorf("failed to authenticate the token, error in status: %s", status.Error)
	}
	user := &models.User{
		Username: status.User.Username,
	}
	for _, au := range adminUsernames {
		if status.User.Username == au {
			log.Debugf("Username: %s in the adminusers list, assigning user admin permission", au)
			user.AdminRoleInAuth = true
			break
		}
	}

	if len(status.User.Groups) > 0 {
		userGroups := model.UserGroupsFromName(status.User.Groups, common.HTTPGroupType)
		groupIDList, err := usergroup.Mgr.Populate(orm.Context(), userGroups)
		if err != nil {
			return nil, err
		}
		log.Debugf("current user's group ID list is %+v", groupIDList)
		user.GroupIDs = groupIDList
		if len(adminGroups) > 0 && !user.AdminRoleInAuth { // skip checking admin group if user already has admin role
			agm := make(map[string]struct{})
			for _, ag := range adminGroups {
				agm[ag] = struct{}{}
			}
			for _, ug := range status.User.Groups {
				if _, ok := agm[ug]; ok {
					user.AdminRoleInAuth = true
					break
				}
			}
		}
	}
	return user, nil
}
