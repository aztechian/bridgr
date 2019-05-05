package workers

// Worker is the interface for how to talk to all instances of worker structs
type Worker interface {
	Setup() error
	Run() error
}
