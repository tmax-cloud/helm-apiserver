package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tmax-cloud/helm-apiserver/pkg/apis"
	"k8s.io/klog"
)

func main() {
	klog.Infoln("initializing server....")

	router := mux.NewRouter()

	// c, err := internal.Client(client.Options{})
	// if err != nil {
	// 	panic(err)
	// }

	router.HandleFunc("/helm", apis.InstallRelease).Methods("POST", "PUT")
	router.HandleFunc("/helm", apis.UnInstallRelease).Methods("DELETE")
	router.HandleFunc("/helm", apis.RollbackRelease).Methods("PATCH")

	http.Handle("/", router)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", 8081), nil); err != nil {
		klog.Errorln(err, "failed to initialize a server")
	}

}
