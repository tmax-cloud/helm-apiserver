package main

import (
	"os"

	"github.com/tmax-cloud/helm-apiserver/internal/hclient"
	"github.com/tmax-cloud/helm-apiserver/pkg/apiserver"

	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog"

	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"

	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	setupLog           = ctrl.Log.WithName("setup")
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion)
	return nil
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(Scheme))
	utilruntime.Must(apiregv1.AddToScheme(Scheme))
	utilruntime.Must(rbac.AddToScheme(Scheme))
	utilruntime.Must(AddToScheme(Scheme))

	// Setting VersionPriority is critical in the InstallAPIGroup call (done in New())
	// utilruntime.Must(Scheme.SetVersionPriority(SchemeGroupVersion))

	// TODO(devdattakulkarni) -- Following comments coming from sample-apiserver.
	// Leaving them for now.
	// we need to add the options to empty v1
	// TODO fix the server code to avoid this
	// metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Group: APIGroup, Version: APIVersion})
	// +kubebuilder:scaffold:scheme
}

func main() {
	klog.Infoln("initializing server....")
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: Scheme,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	hcm := hclient.NewHelmClientManager()

	// Start API aggregation server
	apiServer, err := apiserver.New(mgr.GetClient(), mgr.GetConfig(), hcm, mgr.GetCache())
	if err != nil {
		setupLog.Error(err, "unable to create api server")
		os.Exit(1)
	}
	go apiServer.Start()

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}

}
