package release

import (
	// "bytes"
	// "context"
	// "encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/gorilla/mux"

	// helmclient "github.com/mittwald/go-helm-client"
	mockhelmclient "github.com/mittwald/go-helm-client/mock"
	"github.com/stretchr/testify/assert"
	"github.com/tmax-cloud/helm-apiserver/internal/hclient"

	// "github.com/tmax-cloud/helm-apiserver/pkg/schemas"
	"helm.sh/helm/v3/pkg/release"
)

func TestGetReleases(t *testing.T) {

	hcm := hclient.NewHelmClientManager()

	var releases []*release.Release
	ctrl := gomock.NewController(t)
	m := mockhelmclient.NewMockClient(ctrl)
	m.EXPECT().ListDeployedReleases().Return(releases, nil)

	hcm.Hci = m
	sh := ReleaseHandler{
		hcm: hcm,
	}

	defer ctrl.Finish()

	t.Run("check get releases", func(t *testing.T) {

		req, err := http.NewRequest("GET", "/helm/v1/namespaces/test/releases", nil)
		assert.Nil(t, err, "")
		response := httptest.NewRecorder()
		sh.GetReleases(response, req)

		if status := response.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})
}

// func TestInstallReleases(t *testing.T) {

// 	hcm := hclient.NewHelmClientManager()

// 	chartSpec := helmclient.ChartSpec{
// 		ReleaseName: "test-release",
// 		ChartName:   "test-chart",
// 		ValuesYaml:  "",
// 		Version:     "test",
// 		UpgradeCRDs: true,
// 		Wait:        false,
// 	}

// 	var release *release.Release
// 	ctrl := gomock.NewController(t)
// 	m := mockhelmclient.NewMockClient(ctrl)
// 	m.EXPECT().ListDeployedReleases().Return(nil, nil)
// 	m.EXPECT().InstallOrUpgradeChart(context.Background(), &chartSpec).Return(release, nil)

// 	hcm.Hci = m
// 	sh := ReleaseHandler{
// 		hcm: hcm,
// 	}

// 	defer ctrl.Finish()

// 	t.Run("check install release", func(t *testing.T) {

// 		releaseReqSpec := schemas.ReleaseRequestSpec{
// 			PackageURL:  "test-chart",
// 			ReleaseName: "test-release",
// 			Version:     "test",
// 		}
// 		releaseReq := schemas.ReleaseRequest{
// 			ReleaseRequestSpec: releaseReqSpec,
// 		}
// 		bytereleaseReq, _ := json.Marshal(releaseReq)
// 		reqBody := bytes.NewBuffer(bytereleaseReq)
// 		req, err := http.NewRequest("POST", "/helm/v1/namespaces/test/releases", reqBody)
// 		assert.Nil(t, err, "")
// 		response := httptest.NewRecorder()
// 		sh.InstallRelease(response, req)

// 		if status := response.Code; status != http.StatusOK {
// 			t.Errorf("handler returned wrong status code: got %v want %v",
// 				status, http.StatusOK)
// 		}
// 	})
// }

func TestUnInstallReleases(t *testing.T) {

	hcm := hclient.NewHelmClientManager()

	ctrl := gomock.NewController(t)
	m := mockhelmclient.NewMockClient(ctrl)
	m.EXPECT().UninstallReleaseByName("test-release").Return(nil)

	hcm.Hci = m
	sh := ReleaseHandler{
		hcm: hcm,
	}

	defer ctrl.Finish()

	t.Run("check uninstall release", func(t *testing.T) {

		req, err := http.NewRequest("DELETE", "/helm/v1/namespaces/test/releases", nil)
		assert.Nil(t, err, "")
		req = mux.SetURLVars(req, map[string]string{"release-name": "test-release"})

		response := httptest.NewRecorder()
		sh.UnInstallRelease(response, req)

		if status := response.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})
}
