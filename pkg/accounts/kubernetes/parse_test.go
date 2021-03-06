package kubernetes

import (
	"context"
	"github.com/armory/spinnaker-operator/pkg/accounts/account"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func init() {
	TypesFactory = test.TypesFactory
}

func TestFromCRD(t *testing.T) {
	tests := []struct {
		name     string
		manifest string
		expected func(t *testing.T, a account.Account, err error)
	}{
		{
			name: "no kubernetes section in CRD",
			manifest: `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerAccount
metadata:
  name: account1
spec:
  type: Kubernetes
`,
			expected: func(t *testing.T, _ account.Account, err error) {
				assert.Equal(t, noKubernetesDefinedError, err)
			},
		},
		{
			name: "no kubernetes auth section in CRD",
			manifest: `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerAccount
metadata:
  name: account1
spec:
  type: Kubernetes
  kubernetes: {}
`,
			expected: func(t *testing.T, _ account.Account, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "full kubernetes config",
			manifest: `
apiVersion: spinnaker.io/v1alpha2
kind: SpinnakerAccount
metadata:
  name: account1
spec:
  type: Kubernetes
  kubernetes: {}
  settings:
    name: kubernetes
    requiredGroupMembership: []
    providerVersion: V2
    permissions: {}
    dockerRegistries: []
    configureImagePullSecrets: true
    cacheThreads: 1
    namespaces:
    - ns1
    - ns2
    omitNamespaces: []
    kinds: []
    omitKinds: []
    customResources: []
    cachingPolicies: []
    oAuthScopes: []
    onlySpinnakerManaged: false
`,
			expected: func(t *testing.T, _ account.Account, err error) {
				assert.Nil(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sa := TypesFactory.NewAccount()
			if !assert.Nil(t, yaml.Unmarshal([]byte(tt.manifest), sa)) {
				return
			}
			k := &AccountType{}
			a, err := k.FromCRD(sa)
			tt.expected(t, a, err)
		})
	}
}

func TestFromSpinnakerSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings map[string]interface{}
		expected func(t *testing.T, a account.Account, err error)
	}{
		{
			name: "basic settings with kubeconfigFile",
			settings: map[string]interface{}{
				"name":           "test",
				"kubeconfigFile": "mykubeconfig",
			},
			expected: func(t *testing.T, _ account.Account, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "basic settings with service account",
			settings: map[string]interface{}{
				"name":           "test",
				"serviceAccount": true,
			},
			expected: func(t *testing.T, _ account.Account, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "basic settings with kubeconfig inlined",
			settings: map[string]interface{}{
				"name": "test",
				"kubeconfigContents": `
apiVersion: v1
kind: Config
current-context: test-context
clusters:
- cluster:
    api-version: v1
    server: http://mycluster.com
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
users:
- name: test-user
  user:
    token: test-token
`,
			},
			expected: func(t *testing.T, _ account.Account, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "but names are still required",
			settings: map[string]interface{}{
				"kubeconfigFile": "test",
			},
			expected: func(t *testing.T, _ account.Account, err error) {
				if assert.NotNil(t, err) {
					assert.Equal(t, "Kubernetes account missing name", err.Error())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &AccountType{}
			a, err := k.FromSpinnakerConfig(context.TODO(), tt.settings)
			tt.expected(t, a, err)
		})
	}
}

func TestToSpinnakerSettingsAuth(t *testing.T) {
	tests := []struct {
		name     string
		auth     *interfaces.KubernetesAuth
		expected func(t *testing.T, ss map[string]interface{}, err error)
	}{
		{
			name: "service account auth",
			auth: &interfaces.KubernetesAuth{UseServiceAccount: true},
			expected: func(t *testing.T, ss map[string]interface{}, err error) {
				assert.Nil(t, err)
				assert.True(t, ss[UseServiceAccount].(bool))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Account{
				Name:     "kube-sa",
				Auth:     &interfaces.KubernetesAuth{UseServiceAccount: true},
				Env:      Env{},
				Settings: interfaces.FreeForm{},
			}
			ss, err := k.ToSpinnakerSettings(context.TODO())
			tt.expected(t, ss, err)
		})
	}
}
