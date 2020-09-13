package main

import (
	"github.com/Skactor/mitmproxy/config"
	"github.com/Skactor/mitmproxy/export"
	"github.com/Skactor/mitmproxy/logger"
	"github.com/Skactor/mitmproxy/mitm"
	"github.com/elazarl/goproxy"
	"github.com/vardius/message-bus"
	"log"
	"net/http"
)

func main() {
	var exporter export.Exporter

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
	var bus = messagebus.New(100)
	_ = bus.Subscribe("response", func(resp *http.Response, i int) {
		output, err := export.OutputRequestFromResponse(resp)
		if err != nil {
			logger.Logger.Errorf("Failed to parse response with error: %s", err.Error())
			return
		}
		err = exporter.WriteInterface(output)
		if err != nil {
			logger.Logger.Error(err.Error())
			if i < 2 {
				exporter.Open(cfg.Exporter.Config)
				bus.Publish("response", resp, i+1)
			}
			return
		}
	})

	err = mitm.SetCA(cfg.Server)
	if err != nil {
		log.Fatalf("Failed to set ca: %s", err.Error())
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.OnResponse().DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			bus.Publish("response", resp, 0)
			return resp
		},
	)
	logger.Logger.Infof("Starting mitm proxy server on %s...", cfg.Server.Address)
	logger.Logger.Fatal(http.ListenAndServe(cfg.Server.Address, proxy).Error())
}
