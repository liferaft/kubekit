package stacki

// GetPublicKey return the public key and file from the configuration, also if
// this platform requires a public key for provisioning
func (p *Platform) GetPublicKey() (string, []byte, bool) {
	required := len(p.config.Password) == 0
	return p.config.PublicKeyFile, []byte(p.config.PublicKey), required
}

// PublicKey sets the public key and file in the configuration and variables
func (p *Platform) PublicKey(file string, key []byte) {
	p.config.PublicKeyFile = file
	p.config.PublicKey = string(key)
}

// GetPrivateKey returns the private key and file from the configuration, also
// if this platform requires a private key for provisioning
func (p *Platform) GetPrivateKey() (string, []byte, bool) {
	required := len(p.config.Password) == 0
	return p.config.PrivateKeyFile, []byte(p.config.PrivateKey), required
}

// PrivateKey sets the private key and file in the configuration
func (p *Platform) PrivateKey(file string, encKey, key []byte) {
	p.config.PrivateKeyFile = file
	p.config.PrivateKey = string(encKey)
}

// Credentials is to assign the credentials to the configuration
func (p *Platform) Credentials(params ...string) {
	p.ui.Log.Debugf("%s platform does not implements Credentials()", p.name)
}
