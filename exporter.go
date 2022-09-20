package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dmartinol/deployment-exporter/pkg/builder"
	"github.com/dmartinol/deployment-exporter/pkg/formatter"
	log "github.com/dmartinol/deployment-exporter/pkg/log"
	logger "github.com/dmartinol/deployment-exporter/pkg/log"

	"github.com/gorilla/mux"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	log.InitLogger()
	if _, ok := os.LookupEnv("CONTAINER_MODE"); ok {
		logger.Infof("Running in CONTAINER_MODE")
	} else {
		logger.Infof("Running in LOCAL_MODE")
	}
	startServer()
}

var kubeconfig *string
var router = mux.NewRouter()

func startServer() {
	router.Path("/inventory").Queries("type", "{filter}").HandlerFunc(inventoryHandler).Name("inventoryHandler")
	router.Path("/inventory").HandlerFunc(inventoryHandler).Name("inventoryHandler")

	url := "localhost:" + serverPortOrDefault()
	logger.Infof("Starting listener as %s", url)
	if err := http.ListenAndServe(url, router); err != nil {
		log.Fatal(err)
	}
}

func serverPortOrDefault() string {
	if port, ok := os.LookupEnv("SERVERPORT"); ok {
		return port
	}
	return "8080"
}

func inventoryHandler(rw http.ResponseWriter, req *http.Request) {
	contentType := req.FormValue("type")

	if req.URL.Path == "/inventory" {
		if req.Method == "GET" {
			inventory(contentType, rw)
		} else {
			http.Error(rw, fmt.Sprintf("Expect method GET at /, got %v", req.Method), http.StatusMethodNotAllowed)
		}
	} else {
		http.Error(rw, fmt.Sprintf("Unmanaged path %s", req.URL.Path), http.StatusNotFound)
		return
	}
}

func inventory(contentType string, rw http.ResponseWriter) {
	config, err := connectCluster()
	if err != nil {
		http.Error(rw, fmt.Sprintf("Cannot connect cluster: %s", err), http.StatusInternalServerError)
	}
	log.Info("Cluster connected")

	topology, err := builder.NewModelBuilder().BuildForConfig(config)
	if err != nil {
		http.Error(rw, fmt.Sprintf("Cannot build data model: %s", err), http.StatusInternalServerError)
	}

	fmt := formatter.Formatter{TopologyModel: *topology}
	fmt.Format(formatter.ContentTypeFromString(contentType), rw)
}

func initKubeconfig() *string {
	if home := homeDir(); home != "" {
		return flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "")
	} else {
		return flag.String("kubeconfig", "", "")
	}
}

func connectCluster() (*rest.Config, error) {
	if _, ok := os.LookupEnv("CONTAINER_MODE"); ok {
		return rest.InClusterConfig()
	} else {
		if kubeconfig == nil {
			kubeconfig = initKubeconfig()
		}
		//Load config for Openshift's go-client from kubeconfig file
		return clientcmd.BuildConfigFromFlags("", *kubeconfig)
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
