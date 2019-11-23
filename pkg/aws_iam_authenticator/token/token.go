package token

import (
	"fmt"

	"github.com/kubernetes-sigs/aws-iam-authenticator/pkg/token"
	"github.com/liferaft/kubekit/pkg/kluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientauthv1alpha1 "k8s.io/client-go/pkg/apis/clientauthentication/v1alpha1"
)

// Token encapsulates the token from aws-iam-authenticator and extend its functionality
type Token struct {
	clientauthv1alpha1.ExecCredential
	generator token.Generator
	token     token.Token
}

// type ExecCredential struct {
// 	metav1.TypeMeta `json:",inline"`

// 	// Spec holds information passed to the plugin by the transport. This contains
// 	// request and runtime specific information, such as if the session is interactive.
// 	Spec ExecCredentialSpec `json:"spec,omitempty"`

// 	// Status is filled in by the plugin and holds the credentials that the transport
// 	// should use to contact the API.
// 	// +optional
// 	Status *ExecCredentialStatus `json:"status,omitempty"`
// }

// GenerateToken returns an AWS IAM Authenticator token
func GenerateToken(cluster *kluster.Kluster, roleARN string) (*Token, error) {
	var t token.Token
	var err error

	gen, err := token.NewGenerator(false)
	if err != nil {
		return nil, fmt.Errorf("could not get token: %s", err)
	}

	creds, err := cluster.GetCredentialsAsMap()
	if err != nil {
		return nil, fmt.Errorf("could not get cluster credentials: %s", err)
	}

	sess, err := kluster.GetSession(creds)

	if err == nil {
		t, _ = gen.GetWithRoleForSession(cluster.Name, roleARN, sess)
	}

	if t == (token.Token{}) {
		t, err = gen.GetWithRole(cluster.Name, roleARN)
	}
	if err != nil || t == (token.Token{}) {
		return nil, fmt.Errorf("could not get token: %s", err)
	}

	expirationTimestamp := metav1.NewTime(t.Expiration)
	tkn := &Token{
		generator: gen,
		token:     t,
	}
	tkn.TypeMeta = metav1.TypeMeta{
		APIVersion: "client.authentication.k8s.io/v1alpha1",
		Kind:       "ExecCredential",
	}
	tkn.Status = &clientauthv1alpha1.ExecCredentialStatus{
		ExpirationTimestamp: &expirationTimestamp,
		Token:               t.Token,
	}

	// enc, _ := json.Marshal(execInput)
	// string(enc)

	return tkn, nil
}

// FormatJSON returns the token in JSON format
func (t *Token) FormatJSON() string {
	return t.generator.FormatJSON(t.token)
}
