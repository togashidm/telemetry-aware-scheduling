package main

import (
	"flag"
	"github.com/intel/telemetry-aware-scheduling/pkg/cache"
	"github.com/intel/telemetry-aware-scheduling/pkg/scheduler"
	"github.com/intel/telemetry-aware-scheduling/pkg/telemetryscheduler"
)

func main() {
	var kubeConfig, port, certFile, keyFile, cacheEndpoint string
	var unsafe bool
	flag.StringVar(&kubeConfig, "kubeConfig", "/root/.kube/config", "location of kubernetes config file")
	flag.StringVar(&port, "port", "9001", "port on which the scheduler extender will listen")
	flag.StringVar(&certFile, "cert", "/etc/kubernetes/pki/ca.crt", "cert file extender will use for authentication")
	flag.StringVar(&keyFile, "key", "/etc/kubernetes/pki/ca.key", "key file extender will use for authentication")
	flag.StringVar(&cacheEndpoint, "cacheEndpoint", "http://127.0.0.1:8111/cache/", "root at which the cache can be reached for reading")
	flag.BoolVar(&unsafe, "unsafe", false, "unsafe instances of telemetry aware scheduler will be served over simple http.")
	flag.Parse()
	cacheReader := cache.RemoteClient{}
	cacheReader.RegisterEndpoint(cacheEndpoint)
	tscheduler := telemetryscheduler.NewMetricsExtender(&cacheReader)
	sch := scheduler.Server{ExtenderScheduler: tscheduler}
	sch.StartServer(port, certFile, keyFile, unsafe)
}
