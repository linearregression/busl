package broker

// Registrar is a basic broker interface
type Registrar interface {
	Register(key string) error
	IsRegistered(key string) bool
}
