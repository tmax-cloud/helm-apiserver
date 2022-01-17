package apis

import (
	"encoding/json"
	"net/http"

	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/klog"
)

func AddChartRepo(w http.ResponseWriter, r *http.Request) {

	req := schemas.ReleaseRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		klog.Errorln(err, "failed to decode request")
		return
	}

	chartRepo := repo.Entry{
		Name: "test",
		URL:  req.Spec.Repository,
	}

	settings := cli.New()
	repository, _ := repo.NewChartRepository(&chartRepo, getter.All(settings))
	idxPath, err := repository.DownloadIndexFile()
	if idxPath == "" {
		klog.Errorln(err, "failed to get index file")
	}

	index, _ := repo.LoadIndexFile(idxPath)

	// for test
	for _, entry := range index.Entries {
		for _, chartversion := range entry {
			klog.Infoln("Chart Name :" + chartversion.Name)
			klog.Infoln(chartversion)
		}
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(""); err != nil {
		klog.Errorln(err, "failed to encode response")
	}

}
