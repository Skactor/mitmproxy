package export

type Exporter interface {
	Open(cfg interface{}) error
	WriteBytes(data []byte) error
	WriteInterface(data interface{}) error
	Close() error
}
