package encoders

type Encoder interface {
	Seek(offset int64, whence int) (int64, error)
	Read(p []byte) (int, error)
}
