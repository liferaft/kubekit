module github.com/liferaft/kubekit

go 1.13

require (
	github.com/Azure/azure-sdk-for-go v32.5.0+incompatible
	github.com/Azure/go-autorest/autorest/to v0.3.0
	github.com/aws/aws-sdk-go v1.25.4
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/go-ini/ini v1.48.0
	github.com/golang/protobuf v1.3.2
	github.com/golang/snappy v0.0.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.11.3
	github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/terraform v0.12.12
	github.com/iancoleman/strcase v0.0.0-20190422225806-e506e3ef7365
	github.com/johandry/log v0.0.0-20190918193429-2b13006dd125
	github.com/johandry/merger v0.0.0-20190722191252-46f5dfdce4bd
	github.com/kraken/terraformer v0.2.0
	github.com/kraken/ui v0.0.1
	github.com/kubekit/azure v0.0.11
	github.com/kubernetes-sigs/aws-iam-authenticator v0.3.1-0.20181019024009-82544ec86140
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nightlyone/lockfile v0.0.0-20180618180623-0ad87eef1443
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d
	github.com/pelletier/go-toml v1.4.0
	github.com/pkg/sftp v1.10.1
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	github.com/terraform-providers/terraform-provider-aws v1.60.1-0.20191003145700-f8707a46c6ec
	github.com/terraform-providers/terraform-provider-azurerm v1.34.0
	github.com/terraform-providers/terraform-provider-openstack v1.23.0
	github.com/terraform-providers/terraform-provider-template v1.0.1-0.20190501175038-5333ad92003c
	github.com/terraform-providers/terraform-provider-vsphere v1.13.0
	github.com/zclconf/go-cty v1.1.0
	golang.org/x/crypto v0.0.0-20191029031824-8986dd9e96cf
	golang.org/x/net v0.0.0-20191009170851-d66e71096ffb
	golang.org/x/sys v0.0.0-20191029155521-f43be2a4598c // indirect
	// golang.org/x/crypto v0.0.0-20191002192127-34f69633bfdc
	// golang.org/x/net v0.0.0-20191009170851-d66e71096ffb
	google.golang.org/genproto v0.0.0-20191002211648-c459b9ce5143
	google.golang.org/grpc v1.24.0
	gopkg.in/ini.v1 v1.48.0 // indirect
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/cli-runtime v0.0.0
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kubernetes v1.15.0
)

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.2.0
	github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8

	github.com/kraken/terraformer => ./staging/src/github.com/kraken/terraformer
	github.com/kraken/ui => ./staging/src/github.com/kraken/ui
	github.com/kubekit/azure => ./staging/src/github.com/kubekit/azure
	github.com/terraform-providers/terraform-provider-tls => github.com/terraform-providers/terraform-provider-tls v1.2.1-0.20190816230231-0790c4b40281
	github.com/vmware/vic => github.com/pokstad/vic v1.5.1-alpha

	golang.org/x/tools v0.0.0-20190314010720-f0bfdbff1f9c => golang.org/x/tools v0.0.0-20191009213438-b090f1f24028
	k8s.io/api => k8s.io/api v0.0.0-20190620084959-7cf5895f2711
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190620085554-14e95df34f1f
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190620085212-47dc9a115b18
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190620085706-2090e6d8f84c
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190620090043-8301c0bda1f0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190620090013-c9a0fc045dc1
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190620085130-185d68e6e6ea
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190531030430-6117653b35f1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190620090116-299a7b270edc
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190620085325-f29e2b4a4f84
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190620085942-b7f18460b210
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190620085809-589f994ddf7f
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190620085912-4acac5405ec6
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190620085838-f1cb295a73c9
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190620090156-2138f2c9de18
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190620085625-3b22d835f165
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190620085408-1aef9010884e
	mvdan.cc/unparam v0.0.0-20190124213536-fbb59629db34 => mvdan.cc/unparam v0.0.0-20190917161559-b83a221c10a2
)
