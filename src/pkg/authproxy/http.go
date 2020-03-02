package authproxy

import (
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao/group"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	k8s_api_v1beta1 "k8s.io/api/authentication/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// TokenReview ...
func TokenReview(rawToken string, authProxyConfig *models.HTTPAuthProxy) (k8s_api_v1beta1.TokenReviewStatus, error) {

	emptyStatus := k8s_api_v1beta1.TokenReviewStatus{}
	// Init auth client with the auth proxy endpoint.
	authClientCfg := &rest.Config{
		Host: authProxyConfig.TokenReviewEndpoint,
		ContentConfig: rest.ContentConfig{
			GroupVersion:         &schema.GroupVersion{},
			NegotiatedSerializer: serializer.DirectCodecFactory{CodecFactory: scheme.Codecs},
		},
		BearerToken:     rawToken,
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
	res := authClient.Post().Body(tokenReviewRequest).Do()
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

func getTLSConfig(config *models.HTTPAuthProxy) rest.TLSClientConfig {
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
func UserFromReviewStatus(status k8s_api_v1beta1.TokenReviewStatus) (*models.User, error) {
	if !status.Authenticated {
		return nil, fmt.Errorf("failed to authenticate the token, error in status: %s", status.Error)
	}
	user := &models.User{
		Username: status.User.Username,
	}
	if len(status.User.Groups) > 0 {
		userGroups := models.UserGroupsFromName(status.User.Groups, common.HTTPGroupType)
		groupIDList, err := group.PopulateGroup(userGroups)
		if err != nil {
			return nil, err
		}
		log.Debugf("current user's group ID list is %+v", groupIDList)
		user.GroupIDs = groupIDList
	}
	return user, nil

}
