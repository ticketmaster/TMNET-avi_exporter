package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	hosturl       = flag.String(os.Getenv("AVI_CLUSTER"), "", "AVI Cluster URL.")
	listenAddress = flag.String("web.listen-address", ":8080", "Address to listen on for web interface and telemetry.")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
)

func main() {
	flag.Parse()
	//////////////////////////////////////////////////////////////////////////////
	// Set metrics endpoint.
	//////////////////////////////////////////////////////////////////////////////
	e := NewExporter()
	e.registerGauges()
	http.Handle("/metrics", myPromHTTPHandler(e, prometheus.DefaultGatherer, promhttp.HandlerOpts{}))
	//////////////////////////////////////////////////////////////////////////////
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>AVI Exporter</title></head>
             <body>
             <h1>AVI Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	//////////////////////////////////////////////////////////////////////////////
	// Set service health endpoint.
	//////////////////////////////////////////////////////////////////////////////
	u, err := url.Parse(*hosturl)
	if err != nil {
		log.Print(err)
		os.Exit(-1)
	}
	health := healthcheck.NewHandler()
	var port string
	if u.Port() == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	} else {
		port = u.Port()
	}

	health.AddReadinessCheck(
		"avi-tcp",
		healthcheck.Async(healthcheck.TCPDialCheck(u.Host+":"+port, 50*time.Millisecond), 10*time.Second))

	http.HandleFunc("/live", health.LiveEndpoint)
	http.HandleFunc("/healthz", health.ReadyEndpoint)
	//////////////////////////////////////////////////////////////////////////////
	glog.Infoln("Starting HTTP server on", *listenAddress)
	glog.Exitf(http.ListenAndServe(*listenAddress, nil).Error())
}
