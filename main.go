package main

import (
	"github.com/Skactor/mitmproxy/config"
	"github.com/Skactor/mitmproxy/export"
	"github.com/Skactor/mitmproxy/logger"
	"github.com/Skactor/mitmproxy/mitm"
	"github.com/elazarl/goproxy"
	"github.com/panjf2000/ants/v2"
	"log"
	"net/http"
	"net/http/httputil"
)

var exporter export.Exporter

func main() {

	err := logger.InitLogger()
	if err != nil {
		logger.Logger.Fatalf("Failed to init logger with error", err)
	}
	cfg, err := config.Parse("./config.yaml")
	if err != nil {
		log.Fatalf("Failed to parse config file with error: %s", err.Error())
	}
	switch cfg.Exporter.Type {
	case "tcp":
		exporter = &export.TCPExporter{}
	default:
		logger.Logger.Fatalf("Unknown exporter type %s", cfg.Exporter.Type)
		return
	}
	err = exporter.Open(cfg.Exporter.Config)
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}
	defer exporter.Close()

	pool, _ := ants.NewPoolWithFunc(1024, func(i interface{}) {
		err = exporter.WriteInterface(i)
		if err != nil {
			logger.Logger.Error(err.Error())
			exporter.Open(cfg.Exporter.Config)
		}
	})
	err = mitm.SetCA(cfg.Server)
	if err != nil {
		log.Fatalf("Failed to set ca: %s", err.Error())
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			var err error
			ctx.UserData, err = httputil.DumpRequestOut(req, true)
			if err != nil {
				logger.Logger.Error(err.Error())
			}
			return req, nil
		},
	)
	proxy.OnResponse().DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			logger.Logger.Infof("Req: [%d] [%s] %s", resp.StatusCode, resp.Request.Method, resp.Request.URL.String())
			rawResponse, err := httputil.DumpResponse(resp, true)
			if err != nil {
				logger.Logger.Error(err.Error())
			}

			pool.Invoke(map[string][]byte{
				"request":  ctx.UserData.([]byte),
				"response": rawResponse,
			})
			ctx.UserData = nil
			return resp
		},
	)
	logger.Logger.Noticef("Starting mitm proxy server on %s...", cfg.Server.Address)
	logger.Logger.Fatal(http.ListenAndServe(cfg.Server.Address, proxy).Error())
}
