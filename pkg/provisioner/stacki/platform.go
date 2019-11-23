package stacki

// Name returns the platform name
func (p *Platform) Name() string {
	return p.name
}

// Config returns the default configuration for Stacki
func (p *Platform) Config() interface{} {
	return p.config
}
