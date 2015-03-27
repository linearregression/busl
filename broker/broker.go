package broker

type Registrar interface {
	Register(key string) error
	IsRegistered(key string) bool
}
