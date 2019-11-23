package eks

// Name returns the platform name
func (p *Platform) Name() string {
	return p.name
}

// Config returns the default configuration for AWS
func (p *Platform) Config() interface{} {
	return p.config
}
