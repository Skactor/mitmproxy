package mitm

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/Skactor/mitmproxy/config"
	"github.com/elazarl/goproxy"
	"io/ioutil"
)

func SetCA(config config.ServerConfig) (err error) {
	caCert, err := ioutil.ReadFile(config.CertPath)
	if err != nil {
		return err
	}
	caKey, err := ioutil.ReadFile(config.KeyPath)
	if err != nil {
		return err
	}
	loadedProxyCA, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		return err
	}
	if loadedProxyCA.Leaf, err = x509.ParseCertificate(loadedProxyCA.Certificate[0]); err != nil {
		return err
	}
	goproxy.GoproxyCa = loadedProxyCA
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&loadedProxyCA)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&loadedProxyCA)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&loadedProxyCA)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&loadedProxyCA)}
	return nil
}
