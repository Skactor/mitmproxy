package export

import (
	"encoding/json"
	"github.com/pkg/errors"
	"net"
	"sync"
	"time"
)

type TCPExporterConfig struct {
	Address   string `json:"address"`
	KeepAlive int64  `json:"keepalive"`
	NewLine   bool   `json:"newline"`
}

type TCPExporter struct {
	conn   *net.TCPConn
	config *TCPExporterConfig
}

var mutex sync.Mutex

func (e *TCPExporter) parseConfig(cfg interface{}) error {
	newCfg := TCPExporterConfig{}
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &newCfg)
	if err != nil {
		return errors.New("failed to parse export configuration")
	}
	e.config = &newCfg
	return nil
}

func (e *TCPExporter) Open(cfg interface{}) error {
	var err error
	mutex.Lock()
	defer mutex.Unlock()
	err = e.parseConfig(cfg)
	if err != nil {
		return err
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", e.config.Address)
	if err != nil {
		return errors.Errorf("ResolveTCPAddr failed: %s", err.Error())
	}

	e.conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return errors.Errorf("Dial failed: %s", err.Error())
	}
	if e.config.KeepAlive > 0 {
		err = e.conn.SetKeepAlive(true)
		if err != nil {
			return err
		}
		err = e.conn.SetKeepAlivePeriod(time.Duration(e.config.KeepAlive) * time.Millisecond)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *TCPExporter) WriteBytes(data []byte) error {
	if e.config.NewLine {
		data = append(data, byte('\n'))
	}
	if e.conn == nil {
		return errors.New("Some data is being discarded while connecting")
	}
	_, err := e.conn.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (e *TCPExporter) WriteInterface(data interface{}) error {
	encodedData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return e.WriteBytes(encodedData)

}
func (e *TCPExporter) Close() error {
	return e.conn.Close()
}
