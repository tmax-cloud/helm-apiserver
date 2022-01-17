package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tmax-cloud/helm-apiserver/pkg/apis"
	"k8s.io/klog"
)

const (
	releasePrefix = "/release"
	chartPrefix   = "/chart"
)

func main() {
	klog.Infoln("initializing server....")

	router := mux.NewRouter()

	// c, err := internal.Client(client.Options{})
	// if err != nil {
	// 	panic(err)
	// }

	router.HandleFunc(releasePrefix, apis.InstallRelease).Methods("POST", "PUT")
	router.HandleFunc(releasePrefix, apis.UnInstallRelease).Methods("DELETE")
	router.HandleFunc(releasePrefix, apis.RollbackRelease).Methods("PATCH")
	router.HandleFunc(chartPrefix, apis.AddChartRepo).Methods("GET")

	http.Handle("/", router)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", 8081), nil); err != nil {
		klog.Errorln(err, "failed to initialize a server")
	}

}
