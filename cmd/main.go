package main

import (
	"flag"
	"os"

	"github.com/tmax-cloud/helm-apiserver/internal/hclient"
	"github.com/tmax-cloud/helm-apiserver/pkg/apiserver"

	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog"

	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	apiregv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
)

const (
	APIGroup   = "helmapi.tmax.io"
	APIVersion = "v1"
)

var (
	Scheme             = runtime.NewScheme()
	SchemeBuilder      = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme        = SchemeBuilder.AddToScheme
	SchemeGroupVersion = schema.GroupVersion{Group: APIGroup, Version: APIVersion}
	LogLevel           string
	DefaultRepo        bool
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion)
	return nil
}

func init() {
	flag.StringVar(&LogLevel, "log-level", "INFO", "Log Level; TRACE, DEBUG, INFO, WARN, ERROR, FATAL")
	flag.BoolVar(&DefaultRepo, "add-default-repo", false, "add-default-repo: true, false")
	flag.Parse()
	klog.Infoln("LOG_LEVEL = " + LogLevel)

	if LogLevel == "TRACE" || LogLevel == "trace" {
		LogLevel = "5"
	} else if LogLevel == "DEBUG" || LogLevel == "debug" {
		LogLevel = "4"
	} else if LogLevel == "INFO" || LogLevel == "info" {
		LogLevel = "3"
	} else if LogLevel == "WARN" || LogLevel == "warn" {
		LogLevel = "2"
	} else if LogLevel == "ERROR" || LogLevel == "error" {
		LogLevel = "1"
	} else if LogLevel == "FATAL" || LogLevel == "fatal" {
		LogLevel = "0"
	} else {
		klog.Infoln("Unknown log-level paramater. Set to default level INFO")
		LogLevel = "3"
	}
	klog.InitFlags(nil)
	flag.Set("v", LogLevel)
	flag.Parse()

	utilruntime.Must(clientgoscheme.AddToScheme(Scheme))
	utilruntime.Must(apiregv1.AddToScheme(Scheme))
	utilruntime.Must(rbac.AddToScheme(Scheme))
	utilruntime.Must(AddToScheme(Scheme))

	// Setting VersionPriority is critical in the InstallAPIGroup call (done in New())
	// utilruntime.Must(Scheme.SetVersionPriority(SchemeGroupVersion))
	// +kubebuilder:scaffold:scheme
}

func main() {
	klog.V(3).Info("initializing server....")
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: Scheme,
	})
	if err != nil {
		klog.V(1).Info(err, "unable to start manager")
		os.Exit(1)
	}
	hcm := hclient.NewHelmClientManager()

	// Start API server
	apiServer, err := apiserver.New(mgr.GetClient(), mgr.GetConfig(), hcm, mgr.GetCache(), DefaultRepo)
	if err != nil {
		klog.V(1).Info(err, "unable to create api server")
		os.Exit(1)
	}
	go apiServer.Start()

	klog.V(3).Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		klog.V(1).Info(err, "problem running manager")
		os.Exit(1)
	}

}
