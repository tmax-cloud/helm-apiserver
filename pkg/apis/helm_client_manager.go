package apis

import helmclient "github.com/mittwald/go-helm-client"

type HelmClientManager struct {
	Hc helmclient.Client
}
