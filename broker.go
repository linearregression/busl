package busl

type Broker interface {
	Publish(msg []byte)
	Subscribe() chan []byte
	Unsubscribe(ch []byte)
}
