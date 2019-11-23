package raw

// Variables encapsulate all the TF variables.
type Variables struct{}

// CreateVariables creates the variables from the platform configuration
// If the platform does not have configuration, will use the default one
func (p *Platform) CreateVariables() *Variables {
	p.ui.Log.Debugf("%s platform do not implements CreateVariables() neither Variables", p.name)
	return nil
}
