package aks

import "fmt"

// GetPublicKey return the public key and file from the configuration, also if
// this platform requires a public key for provisioning
func (p *Platform) GetPublicKey() (string, []byte, bool) {
	return p.config.PublicKeyFile, []byte(p.config.PublicKey), true
}

// PublicKey sets the public key and file in the configuration and variables
func (p *Platform) PublicKey(file string, key []byte) {
	p.config.PublicKeyFile = file
	p.config.PublicKey = string(key)
}

// GetPrivateKey returns the private key and file from the configuration, also
// if this platform requires a private key for provisioning
func (p *Platform) GetPrivateKey() (string, []byte, bool) {
	return p.config.PrivateKeyFile, []byte(p.config.PrivateKey), true
}

// PrivateKey sets the private key and file in the configuration
func (p *Platform) PrivateKey(file string, encKey, key []byte) {
	p.config.PrivateKeyFile = file
	p.config.PrivateKey = string(encKey)
}

// Credentials is to assign the credentials to the configuration
func (p *Platform) Credentials(params ...string) {
	if len(params) != 4 {
		panic(fmt.Sprintf("received %d credential parameters, expected 4", len(params)))
	}
	p.ui.Log.Debug("getting Azure credentials")
	// To config
	p.config.SubscriptionID = params[0]
	p.config.TenantID = params[1]
	p.config.ClientID = params[2]
	p.config.ClientSecret = params[3]
}
