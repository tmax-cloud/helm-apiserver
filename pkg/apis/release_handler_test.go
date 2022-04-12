package apis

import (
	"net/http"
	"net/http/httptest"
	"testing"

	// helmclient "github.com/mittwald/go-helm-client"
	gomock "github.com/golang/mock/gomock"
	mockhelmclient "github.com/mittwald/go-helm-client/mock"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/release"
	// "helm.sh/helm/v3/pkg/release"
	// "github.com/tmax-cloud/helm-apiserver/internal"
)

func TestGetReleases(t *testing.T) {

	var releases []*release.Release
	ctrl := gomock.NewController(t)
	m := mockhelmclient.NewMockClient(ctrl)
	m.EXPECT().ListDeployedReleases().Return(releases, nil)
	hcm := HelmClientManager{
		Hci: m,
	}

	defer ctrl.Finish()

	// byteRepoReq, _ := json.Marshal(repoReq)
	// reqBody := bytes.NewBuffer(byteRepoReq)

	t.Run("check get releases", func(t *testing.T) {

		req, err := http.NewRequest("GET", "/helm/ns/test/releases", nil)
		assert.Nil(t, err, "")
		response := httptest.NewRecorder()
		hcm.GetReleases(response, req)

		if status := response.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})
}
