package export

import (
	"bytes"
	"net/http"
	"net/http/httputil"
	"test/logger"
)

type Exporter interface {
	Open(cfg interface{}) error
	WriteBytes(data []byte) error
	WriteInterface(data interface{}) error
	Close() error
}

type OutputRequest struct {
	RawRequestHeader  []byte `json:"raw_request"`
	RawRequestBody    []byte `json:"raw_request_body"`
	RawResponseHeader []byte `json:"raw_response"`
	RawResponseBody   []byte `json:"raw_response_body"`
}

func OutputRequestFromResponse(resp *http.Response) (*OutputRequest, error) {
	out := OutputRequest{}
	rawRequest, err := httputil.DumpRequestOut(resp.Request, true)
	if err != nil {
		logger.Logger.Error(err.Error())
		return nil, err
	}
	splitRequest := bytes.SplitN(rawRequest, []byte{13, 10, 13, 10}, 2)
	out.RawRequestHeader, out.RawRequestBody = splitRequest[0], splitRequest[1]

	rawResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		logger.Logger.Error(err.Error())
		return nil, err
	}
	splitResponse := bytes.SplitN(rawResponse, []byte{13, 10, 13, 10}, 2)
	out.RawResponseHeader, out.RawResponseBody = splitResponse[0], splitResponse[1]
	return &out, nil
}
