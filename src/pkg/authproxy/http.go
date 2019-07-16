package authproxy

import (
	"encoding/json"
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
func TokenReview(sessionID string, authProxyConfig *models.HTTPAuthProxy) (*k8s_api_v1beta1.TokenReview, error) {

	// Init auth client with the auth proxy endpoint.
	authClientCfg := &rest.Config{
		Host: authProxyConfig.TokenReviewEndpoint,
		ContentConfig: rest.ContentConfig{
			GroupVersion:         &schema.GroupVersion{},
			NegotiatedSerializer: serializer.DirectCodecFactory{CodecFactory: scheme.Codecs},
		},
		BearerToken: sessionID,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: !authProxyConfig.VerifyCert,
		},
	}
	authClient, err := rest.RESTClientFor(authClientCfg)
	if err != nil {
		return nil, err
	}

	// Do auth with the token.
	tokenReviewRequest := &k8s_api_v1beta1.TokenReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TokenReview",
			APIVersion: "authentication.k8s.io/v1beta1",
		},
		Spec: k8s_api_v1beta1.TokenReviewSpec{
			Token: sessionID,
		},
	}
	res := authClient.Post().Body(tokenReviewRequest).Do()
	err = res.Error()
	if err != nil {
		log.Errorf("fail to POST auth request, %v", err)
		return nil, err
	}
	resRaw, err := res.Raw()
	if err != nil {
		log.Errorf("fail to get raw data of token review, %v", err)
		return nil, err
	}
	// Parse the auth response, check the user name and authenticated status.
	tokenReviewResponse := &k8s_api_v1beta1.TokenReview{}
	err = json.Unmarshal(resRaw, &tokenReviewResponse)
	if err != nil {
		log.Errorf("fail to decode token review, %v", err)
		return nil, err
	}
	return tokenReviewResponse, nil

}
