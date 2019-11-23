package kluster

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/kraken/ui"
	"github.com/kubekit/kubekit/pkg/crypto/tls"
)

// CACert encapsulate the CN and CLI description for a CA cert
type CACert struct {
	CN   string
	Desc string
}

// GenericKeyPairName is the name of the generic key pair
const (
	GenericKeyPairName = "root_ca"
	APIServerCertName  = "node" // TODO: change it in the future to 'kube-node'
)

// CACertNames is a list of CA certificates (key and cert) the user can provide
// with the flags X-ca-{key,cert}-file.
// The certificate with key "root_ca" is the generic one. This is the CA
// certificate to use when the specific one is not given. i.e. if the API server
// certificate is not provided, KubeKit will use the generic cert. If the
// generic one is not provided, then the CA certificate will be self-signed.
// The description is used for the CLI flag 'X-ca-{key,cert}-file'
var CACertNames = map[string]CACert{
	GenericKeyPairName: {CN: "kube-ca", Desc: "used to generate the server API certificate and also it's the generic one used by the non provided certificates"},
	"etcd_root_ca":     {CN: "etcd-ca", Desc: "used to generate the etcd certificates"},
	"ingress_root_ca":  {CN: "ingress-ca", Desc: "used to generate the ingress certificates"},
	"srv_acc":          {CN: "", Desc: ""},
	// "kube_root_ca":     {CN: "kube-ca", Desc: "used to generate the server API certificates"},
}

// Cert encapsulate the CN and CA for a signed cert
type Cert struct {
	CN          string
	O           string
	FromCA      string
	DNSNames    []string
	IPAddresses []string
	ExtKeyUsage []x509.ExtKeyUsage
}

// CertNames is a list of certificates to generate. Initially contain all the
// generated certificates but in 'cmd/configure.go' the certificates given
// by the user are included. If you which to add more certs, add it in
// cmd.CertNames at 'cmd/configure.go' and make sure the name does not contain 'ca'
var CertNames = map[string]Cert{
	APIServerCertName: Cert{
		CN:     "kube-apiserver",
		FromCA: GenericKeyPairName,
		// TODO: The 'node' certificates are generated from 'root_ca' but should be from 'kube_root_ca'.
		// This will be replaced when the cluster works with 'kube_root_ca'
		DNSNames: []string{
			"kubernetes",
			"kubernetes.default",
			"kubernetes.default.svc",
			"kubernetes.default.svc.cluster.local",
			"{{ registry }}",
			"{{ masters }}",
			"{{ ALB }}",
		},
		IPAddresses: []string{
			"172.21.0.1",
			"{{ VIP }}",
			"{{ masters }}",
		},
	},
	"kubelet": Cert{
		CN:     "system:node:{{ hostname }}",
		O:      "system:nodes",
		FromCA: GenericKeyPairName,
		DNSNames: []string{
			"kubernetes",
			"kubernetes.default",
			"kubernetes.default.svc",
			"kubernetes.default.svc.cluster.local",
			"{{ registry }}",
			"{{ masters }}",
			"{{ workers }}",
			"{{ ALB }}",
		},
		IPAddresses: []string{
			"172.21.0.1",
			"{{ VIP }}",
			"{{ masters }}",
			"{{ workers }}",
		},
	},
	"kube_proxy": Cert{
		CN:     "kube-proxy",
		FromCA: GenericKeyPairName,
		DNSNames: []string{
			"kubernetes",
			"kubernetes.default",
			"kubernetes.default.svc",
			"kubernetes.default.svc.cluster.local",
			"{{ registry }}",
			"{{ masters }}",
			"{{ workers }}",
			"{{ ALB }}",
		},
		IPAddresses: []string{
			"172.21.0.1",
			"{{ VIP }}",
			"{{ masters }}",
			"{{ workers }}",
		},
	},
	"kube_controller": Cert{
		CN:     "kube-controller-manager",
		FromCA: GenericKeyPairName,
		DNSNames: []string{
			"kubernetes",
			"kubernetes.default",
			"kubernetes.default.svc",
			"kubernetes.default.svc.cluster.local",
			"{{ registry }}",
			"{{ masters }}",
			"{{ ALB }}",
		},
		IPAddresses: []string{
			"172.21.0.1",
			"{{ VIP }}",
			"{{ masters }}",
		},
	},
	"kube_scheduler": Cert{
		CN:     "kube-scheduler",
		FromCA: GenericKeyPairName,
		DNSNames: []string{
			"kubernetes",
			"kubernetes.default",
			"kubernetes.default.svc",
			"kubernetes.default.svc.cluster.local",
			"{{ registry }}",
			"{{ masters }}",
			"{{ ALB }}",
		},
		IPAddresses: []string{
			"172.21.0.1",
			"{{ VIP }}",
			"{{ masters }}",
		},
	},
	"admin": Cert{
		CN:     "kube-apiserver",
		O:      "system:masters",
		FromCA: GenericKeyPairName,
		DNSNames: []string{
			"kubernetes",
			"kubernetes.default",
			"kubernetes.default.svc",
			"kubernetes.default.svc.cluster.local",
			"{{ registry }}",
			"{{ masters }}",
			"{{ workers }}",
			"{{ ALB }}",
		},
		IPAddresses: []string{
			"172.21.0.1",
			"{{ VIP }}",
			"{{ masters }}",
			"{{ workers }}",
		},
	},
	"etcd_node": Cert{
		CN:     "etcd",
		FromCA: "etcd_root_ca",
		DNSNames: []string{
			"kubernetes",
			"kubernetes.default",
			"kubernetes.default.svc",
			"kubernetes.default.svc.cluster.local",
			"{{ registry }}",
			"{{ masters }}",
			"{{ ALB }}",
		},
		IPAddresses: []string{
			"172.21.0.1",
			"{{ VIP }}",
			"{{ masters }}",
		},
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
	},
	"ingress": Cert{
		CN:     "ingress",
		FromCA: "ingress_root_ca",
		DNSNames: []string{
			"{{ masters }}",
			"{{ workers }}",
			"{{ ingress_additional_dns_alt_names }}", // TODO: Not defined
		},
		IPAddresses: []string{
			"{{ masters }}",
			"{{ workers }}",
			"{{ ingress_additional_dns_alt_ips }}", // TODO: Not defined
		},
	},
	"opa": Cert{
		CN:     "opa",
		FromCA: GenericKeyPairName,
		DNSNames: []string{
			"opa",
			"opa.opa",
			"opa.opa.svc",
			"opa.opa.svc.cluster.local",
			"{{ masters }}",
			"{{ workers }}",
		},
		IPAddresses: []string{
			"172.21.0.1",
			"{{ masters }}",
			"{{ workers }}",
		},
	},
	// "srv_acc": Cert{
	// 	CN:     "srv_acc",
	// 	FromCA: "",
	// },
}

// GenerateCerts generates all the certificates self-signed if the CA Key and
// Cert are not provided.
func (k *Kluster) GenerateCerts(userCACertsFiles tls.KeyPairs, overwrite bool) error {
	platformName := k.Platform()
	logPrefix := fmt.Sprintf("Certificates [ %s@%s ]", k.Name, platformName)
	k.ui.SetLogPrefix(logPrefix)

	k.LoadState()
	k.ui.Log.Debug("state(s) loaded")

	// Create the certificates directory to save them all
	baseCertsDir, err := k.makeCertDir(platformName)
	if err != nil {
		return err
	}

	// delete(userCACertsFiles, GenericKeyPairName)
	// TODO: At this time, do generate 'root_ca' CA certs are they are needed for 'node's certs
	// delete(CACertNames, GenericKeyPairName)

	k.certificates = make(tls.KeyPairs, len(CACertNames)+len(CertNames))

	// Let's create the CA certificates ...
	if _, err := k.genCACertificates(baseCertsDir, userCACertsFiles, platformName); err != nil {
		return err
	}
	// ... the Certificates
	_, err = k.genCertificates(baseCertsDir, platformName)
	return err

	// return k.certificates.Save(overwrite)
}

func (k *Kluster) genCACertificates(baseCertsDir string, userCACertsFiles tls.KeyPairs, platform string) (tls.KeyPairs, error) {
	// TODO: If files are empty (not given by user in flags) check if the files
	// are in the cluster certificates directory

	certificates := make(tls.KeyPairs, len(CACertNames))
	genericKeyPair := userCACertsFiles[GenericKeyPairName]

	existFile := func(filename string) bool {
		if _, err := os.Stat(filename); err != nil {
			if os.IsNotExist(err) {
				return false
			}
		}
		return true
	}

	getCACertSourceFiles := func(kp *tls.KeyPair) (*tls.KeyPair, error) {
		// If the user enter a cert file but not a key file, don't use the cert file
		if kp.CertFile != "" && kp.KeyFile == "" {
			k.ui.Log.Warnf("cannot use the CA Certificate file %s because there is no CA Key for it", kp.CertFile)
			kp.CertFile = ""
		}

		// Option #1: Use the files from the user input if the key file exists.
		// If the key file doesn't exists, this is a user input error and fail
		// If the cert file is empty, will be created.
		// If the cert files doesn't exists, this is a user input error and fail
		if kp.KeyFile != "" {
			if !existFile(kp.KeyFile) {
				return nil, fmt.Errorf("the key file %s for %s does not exists", kp.KeyFile, kp.Name)
			}
			if kp.CertFile != "" && !existFile(kp.CertFile) {
				return nil, fmt.Errorf("the cert file %s for %s does not exists", kp.CertFile, kp.Name)
			}
			k.ui.Log.Infof("using CA certificate for %s from key file %s", kp.Name, kp.KeyFile)
			return &tls.KeyPair{
				CertFile: kp.CertFile,
				KeyFile:  kp.KeyFile,
			}, nil
		}

		// Option #2: Use a previous key file.
		// The key file is empty, if a previous file exists, use it

		// Get the standard filename for the keys
		kp.NewFilenames(baseCertsDir, "")

		// If the key file exists, use it. If the cert file does not exists will be created
		if existFile(kp.KeyFile) {
			k.ui.Log.Debugf("found the key file %s for %s", kp.KeyFile, kp.Name)
			if !existFile(kp.CertFile) {
				k.ui.Log.Warnf("the cert file %s for %s was not found, it will be generated", kp.CertFile, kp.Name)
				kp.CertFile = ""
			}
			k.ui.Log.Infof("using CA certificate for %s from key file %s", kp.Name, kp.KeyFile)
			return &tls.KeyPair{
				CertFile: kp.CertFile,
				KeyFile:  kp.KeyFile,
			}, nil
		}

		// Option #3: Use the generic key, if this is not the generic one
		// The key file is empty and there isn't a previous file, so use the generic key
		// If this is the generic key and you get here, then create it returning empty files

		// If this is the generic key, create it
		if kp.Name == GenericKeyPairName {
			return &tls.KeyPair{
				CertFile: "",
				KeyFile:  "",
			}, nil
		}

		k.ui.Log.Warnf("using generic CA certificates %q as CA certificate for %s", GenericKeyPairName, kp.Name)

		return genericKeyPair, nil
	}

	generateCAKeyPair := func(name string) error {
		certFile, ok := userCACertsFiles[name]
		if !ok {
			return fmt.Errorf("CA certificate information not found for %s", name)
		}

		fromCAKeyPair, err := getCACertSourceFiles(certFile)
		if err != nil {
			return err
		}

		if fromCAKeyPair.CertFile == "" {
			k.ui.Log.Warnf("CA certificate for %s will be self-signed", name)
		}

		if fromCAKeyPair.KeyFile != "" && fromCAKeyPair.CertFile != "" {
			k.ui.Log.Infof("loading CA certificate for %s", name)
		} else {
			k.ui.Log.Infof("generating CA certificate for %s", name)
		}

		kp, err := tls.NewCAKeyPair(fromCAKeyPair, baseCertsDir, name, CACertNames[name].CN)
		if err != nil {
			return err
		}

		k.certificates[name] = kp
		certificates[name] = kp

		// Always overwrite, the use of the flag --generate-certs should be different
		k.ui.Log.Infof("saving private key of certificate %s", name)
		certificates[name].SavePrivateKey(true)
		k.ui.Log.Infof("saving public key of certificate %s", name)
		certificates[name].SaveCertificate(true)

		return nil
	}

	k.ui.Notify("kubernetes", "certificates", "<certificates>", "", ui.Create)
	defer k.ui.Notify("kubernetes", "certificates", "</certificates>", "")

	// First, generate the generic key
	err := generateCAKeyPair(GenericKeyPairName)
	if err != nil {
		return nil, err
	}

	neededForPlatform := func(name string) bool {
		// Only consider the cases where a CA Cert is NOT needed for the platform in place
		switch platform {
		case "eks", "aks":
			switch name {
			case "ingress_root_ca":
			default:
				return false
			}
		}

		return true
	}

	for name := range CACertNames {
		if name == GenericKeyPairName {
			continue
		}

		if !neededForPlatform(name) {
			continue
		}

		if err := generateCAKeyPair(name); err != nil {
			return nil, err
		}
	}

	return certificates, nil
}

func (k *Kluster) genCertificates(baseCertsDir string, platform string) (tls.KeyPairs, error) {
	certificates := make(tls.KeyPairs, len(CertNames))

	clusterIPS := make(map[string][]string, 0)
	clusterDNSs := make(map[string][]string, 0)

	clusterIPS["masters"] = []string{}
	clusterDNSs["masters"] = []string{}
	// Adding workers too, in case they are needed in the future
	clusterIPS["workers"] = []string{}
	clusterDNSs["workers"] = []string{}

	clusterHostnames := []string{}

	nodes := k.State[platform].Nodes
	for i, node := range nodes {
		if node.PrivateIP != "" {
			clusterIPS[node.RoleName+"s"] = append(clusterIPS[node.RoleName+"s"], node.PrivateIP)
		}
		if node.PublicIP != "" {
			clusterIPS[node.RoleName+"s"] = append(clusterIPS[node.RoleName+"s"], node.PublicIP)
		}

		privHostname := strings.Split(node.PrivateDNS, ".")[0]
		pubHostname := strings.Split(node.PublicDNS, ".")[0]
		clusterHostname := fmt.Sprintf("node-%d", i)
		if privHostname != "" {
			clusterDNSs[node.RoleName+"s"] = append(clusterDNSs[node.RoleName+"s"], node.PrivateDNS, privHostname)
			clusterHostname = privHostname
		}
		if pubHostname != "" {
			clusterDNSs[node.RoleName+"s"] = append(clusterDNSs[node.RoleName+"s"], node.PublicDNS, pubHostname)
			clusterHostname = pubHostname
		}

		clusterHostnames = append(clusterHostnames, clusterHostname)
	}

	address := k.State[platform].Address
	if len(address) > 0 {
		if isIP, _ := regexp.MatchString(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`, address); isIP {
			clusterIPS["VIP"] = []string{address}
		} else {
			clusterDNSs["ALB"] = []string{address}
		}
	}

	// TODO: Cleanup nodeIP and nodeDNS to have uniq elements?

	r, _ := regexp.Compile("{{ *([^ ]+) *}}")

	// equalSlice := func(a, b []string) bool {
	// 	if len(a) != len(b) {
	// 		return false
	// 	}
	// 	var found bool
	// 	for _, ae := range a {
	// 		found = false
	// 		for _, be := range b {
	// 			if ae == be {
	// 				found = true
	// 				break
	// 			}
	// 		}
	// 		if !found {
	// 			return false
	// 		}
	// 	}

	// 	return true
	// }

	// equal := func(kp *tls.KeyPair, cert Cert) bool {
	// 	if kp.IsCA {
	// 		k.ui.Log.Warnf("certificate for %s does not have attribute %q equal as the default configuration. (%v != %v)", kp.Name, "IsCA", kp.IsCA, "false")
	// 		return false
	// 	}
	// 	if kp.CN != cert.CN {
	// 		k.ui.Log.Warnf("certificate for %s does not have attribute %q equal as the default configuration. (%s != %s)", kp.Name, "CN", kp.CN, cert.CN)
	// 		return false
	// 	}
	// 	if kp.O != cert.O {
	// 		k.ui.Log.Warnf("certificate for %s does not have attribute %q equal as the default configuration. (%s != %s)", kp.Name, "O", kp.O, cert.O)
	// 		return false
	// 	}
	// 	// Do not compare DNSNames because cert contain template values (i.e. {{ workers }}) while kp contain the actual values.
	// 	if !equalSlice(kp.DNSNames, cert.DNSNames) {
	// 		k.ui.Log.Warnf("certificate for %s does not have attribute %q equal as the default configuration. (%v != %v)", kp.Name, "DNSNames", kp.DNSNames, cert.DNSNames)
	// 		return false
	// 	}
	// 	// Do not compare IPAddresses because cert contain template values (i.e. {{ masters }}) while kp contain the actual values.
	// 	if !equalSlice(kp.IPAddresses, cert.IPAddresses) {
	// 		k.ui.Log.Warnf("certificate for %s does not have attribute %q equal as the default configuration. (%v != %v)", kp.Name, "IPAddresses", kp.IPAddresses, cert.IPAddresses)
	// 		return false
	// 	}

	// 	// TODO: Compare ExtKeyUsage

	// 	return true
	// }

	saveKeyPair := func(name string, kp *tls.KeyPair) error {
		k.certificates[name] = kp
		certificates[name] = kp

		// Always overwrite, the use of the flag --generate-certs should be different
		k.ui.Log.Infof("saving private key of certificate %s", name)
		if err := certificates[name].SavePrivateKey(true); err != nil {
			return err
		}
		k.ui.Log.Infof("saving public key of certificate %s", name)
		return certificates[name].SaveCertificate(true)
	}

	getKeyPair := func(name string, certInfo Cert, baseCertsDir, hostname string, dns, ips []string) error {
		// If this is a new certificate to create:
		var fromCACertText string
		if len(certInfo.FromCA) != 0 {
			fromCACertText = fmt.Sprintf(" from CA certificate %s", certInfo.FromCA)
		}

		certName := name
		cn := certInfo.CN
		loadedMsg := fmt.Sprintf("loaded certificate for %s", name)
		genMsg := fmt.Sprintf("generating the certificate for %s%s", name, fromCACertText)

		if len(hostname) != 0 {
			certName = name + "@" + hostname
			cn = strings.Replace(certInfo.CN, "{{ hostname }}", hostname, 1)
			loadedMsg = fmt.Sprintf("loaded the node %s certificate for %s", hostname, name)
			genMsg = fmt.Sprintf("generating the node %s certificate for %s%s", hostname, name, fromCACertText)
		}

		if kp, err := tls.Load(baseCertsDir, name, cn); err == nil {
			k.ui.Log.Infof(loadedMsg)
			// if !equal(kp, certInfo) {
			// 	k.ui.Log.Warnf("loaded certificate for %s, does not contain the exact default configuration", name)
			// }

			return saveKeyPair(certName, kp)
		}

		k.ui.Log.Infof(genMsg)
		// k.ui.Log.Debugf("using the IPs %s and DNSs %s", ips, dns)
		kp, err := tls.NewKeyPair(baseCertsDir, name, cn, certInfo.O, dns, ips, k.certificates[certInfo.FromCA], certInfo.ExtKeyUsage)
		if err != nil {
			return err
		}

		return saveKeyPair(certName, kp)
	}

	neededForPlatform := func(name string) bool {
		// Only consider the cases where a CA Cert is NOT needed for the platform in place
		switch platform {
		case "aks":
			switch name {
			case "ingress":
			default:
				return false
			}
		case "eks":
			switch name {
			case "ingress", "opa":
			default:
				return false
			}
		}

		return true
	}

	for name, certInfo := range CertNames {
		// TODO: If the certificates may come from the user:

		if !neededForPlatform(name) {
			continue
		}

		ips := tls.GenericIPAddresses
		for _, ip := range certInfo.IPAddresses {
			matches := r.FindStringSubmatch(ip)
			if len(matches) > 1 {
				if ips2add, ok := clusterIPS[matches[1]]; ok {
					ips = append(ips, ips2add...)
				}
				continue
			}
			ips = append(ips, ip)
		}

		dns := tls.GenericDNSNames
		for _, dn := range certInfo.DNSNames {
			matches := r.FindStringSubmatch(dn)
			if len(matches) > 1 {
				if dns2add, ok := clusterDNSs[matches[1]]; ok {
					dns = append(dns, dns2add...)
				}
				continue
			}
			dns = append(dns, dn)
		}

		if strings.Contains(certInfo.CN, "{{ hostname }}") {
			for _, hostname := range clusterHostnames {
				baseCertsHostDir := filepath.Join(baseCertsDir, hostname)
				if err := os.MkdirAll(baseCertsHostDir, 0700); err != nil {
					return certificates, err
				}

				if err := getKeyPair(name, certInfo, baseCertsHostDir, hostname, dns, ips); err != nil {
					return certificates, err
				}
			}

			continue
		}

		if err := getKeyPair(name, certInfo, baseCertsDir, "", dns, ips); err != nil {
			return certificates, err
		}
	}

	return certificates, nil
}

// GenerateKubeConfig generates the KubeConfig file for a cluster
func (k *Kluster) GenerateKubeConfig() ([]byte, error) {
	platform := k.Platform()

	var tmpl string
	var tmplData KubeconfigData

	switch platform {
	case "aks":
		// AKS provides the kubeconfig, so there is no need to fill out a template
		p, ok := k.provisioner[platform]
		if !ok {
			return nil, fmt.Errorf("AKS platform is not yet a provisioner")
		}

		return []byte(p.Output("kubeconfig")), nil
	case "eks":
		tmpl = KubeconfigEKSTmpl
		p, ok := k.provisioner[platform]
		if !ok {
			return nil, fmt.Errorf("EKS platform is not yet a provisioner")
		}
		tmplData = KubeconfigData{
			CertificateAuthorityData: p.Output("certificate-authority"),
			ClusterName:              k.Name,
			Server:                   k.State[platform].Address,
		}
	default:
		tmpl = KubeconfigTmpl
		APIServerCAName := CertNames["admin"].FromCA
		tmplData = KubeconfigData{
			ClusterName:              k.Name,
			Server:                   k.State[platform].Address,
			Port:                     k.State[platform].Port,
			CertificateAuthorityData: base64.StdEncoding.EncodeToString(k.certificates[APIServerCAName].CertificatePEM),
			ClientCertData:           base64.StdEncoding.EncodeToString(k.certificates["admin"].CertificatePEM),
			ClientKeyData:            base64.StdEncoding.EncodeToString(k.certificates["admin"].PrivateKeyPEM),
		}
	}

	var content bytes.Buffer
	kubeconfigTmpl := template.Must(template.New("kubeconfig").Parse(tmpl))
	err := kubeconfigTmpl.Execute(&content, tmplData)

	return content.Bytes(), err
}

// WriteKubeConfig saves the kubeconfig content in a file into the cluster directory
func (k *Kluster) WriteKubeConfig(kubeconfigContent []byte) (string, error) {
	baseCertsDir, err := k.makeCertDir()
	if err != nil {
		return "", err
	}

	kubeconfigFilename := filepath.Join(baseCertsDir, "kubeconfig")
	kubeconfigFile, err := os.Create(kubeconfigFilename)
	if err != nil {
		return "", fmt.Errorf("failed to create the KubeConfig file %s. %s", kubeconfigFilename, err)
	}
	defer kubeconfigFile.Close()

	return kubeconfigFilename, ioutil.WriteFile(kubeconfigFilename, kubeconfigContent, 0644)
}

// CreateKubeConfigFile creates the kubeconfig file for this cluster
func (k *Kluster) CreateKubeConfigFile() error {
	content, err := k.GenerateKubeConfig()
	if err != nil {
		return err
	}
	k.ui.Notify("kubernetes", "kubeconfig", "<kubeconfig>", "", ui.Create)
	defer k.ui.Notify("kubernetes", "kubeconfig", "</kubeconfig>", "")

	var kubeconfigFilename string
	if kubeconfigFilename, err = k.WriteKubeConfig(content); err != nil {
		return err
	}
	k.ui.Log.Infof("generated kubeconfig to %s", kubeconfigFilename)
	return nil
}

// KubeconfigData contain the data to render the kubeconfig template
type KubeconfigData struct {
	CertificateAuthorityData string
	ClusterName              string
	Server                   string
	Port                     int
	ClientCertData           string
	ClientKeyData            string
}

// KubeconfigTmpl is the kubeconfig template
const KubeconfigTmpl = `
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: |-
      {{ .CertificateAuthorityData }}
    server: https://{{ .Server }}:{{ .Port }}
  name: {{ .ClusterName }}
contexts:
- context:
    cluster: {{ .ClusterName }}
    user:  {{ .ClusterName }}-admin
  name:  {{ .ClusterName }}
current-context:  {{ .ClusterName }}
kind: Config
preferences: {}
users:
- name:  {{ .ClusterName }}-admin
  user:
    client-certificate-data: |-
      {{ .ClientCertData }}
    client-key-data: |-
      {{ .ClientKeyData }}
`

// KubeconfigEKSTmpl is the kubeconfig template for EKS
const KubeconfigEKSTmpl = `
apiVersion: v1
clusters:
- cluster:
    server: {{ .Server }}
    certificate-authority-data: {{ .CertificateAuthorityData }}
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: aws
  name: aws
current-context: aws
kind: Config
preferences: {}
users:
- name: aws
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1alpha1
      command: kubekit
      args:
        - "token"
        - "-i"
        - "{{ .ClusterName }}"
`
