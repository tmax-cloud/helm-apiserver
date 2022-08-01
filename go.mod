module github.com/tmax-cloud/helm-apiserver

go 1.16

require (
	github.com/ghodss/yaml v1.0.0
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/golang/mock v1.6.0
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.5.0
	github.com/jinzhu/copier v0.3.5
	github.com/klauspost/compress v1.15.6 // indirect
	github.com/lib/pq v1.10.6
	github.com/mattn/go-sqlite3 v1.14.13 // indirect
	github.com/mittwald/go-helm-client v0.8.2
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.1
	github.com/xdg-go/scram v1.1.1 // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	go.mongodb.org/mongo-driver v1.9.1
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f // indirect
	gorm.io/driver/sqlite v1.3.4
	gorm.io/gorm v1.23.5
	helm.sh/helm/v3 v3.7.1
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/cli-runtime v0.22.1
	k8s.io/client-go v0.22.2
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.10.3
)
