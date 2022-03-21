module github.com/tmax-cloud/helm-apiserver

go 1.16

require (
	github.com/ghodss/yaml v1.0.0
	github.com/gorilla/mux v1.8.0
	github.com/jinzhu/copier v0.3.5
	github.com/mittwald/go-helm-client v0.8.2
	github.com/spf13/pflag v1.0.5
	helm.sh/helm/v3 v3.7.1
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/cli-runtime v0.22.1
	k8s.io/client-go v0.22.2
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.10.3
)
