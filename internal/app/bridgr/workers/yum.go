package workers

import "bridgr/internal/app/bridgr/config"

// Yum is the worker implementation for Yum repositories
type Yum struct{}

// Run sets up, creates and fetches a YUM repository based on the settings from the config file
func (Yum) Run(conf config.BridgrConf) error {
	return nil
}

// Setup only does the setup step of the YUM worker
func (Yum) Setup(conf config.BridgrConf) error {
	return nil
}
