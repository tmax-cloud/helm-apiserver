package apis

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"helm.sh/helm/v3/pkg/repo"

	// helmclient "github.com/mittwald/go-helm-client"
	gomock "github.com/golang/mock/gomock"
	mockhelmclient "github.com/mittwald/go-helm-client/mock"
	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
	// "github.com/tmax-cloud/helm-apiserver/internal"
)

// var (
// 	server  *httptest.Server
// 	testUrl string
// )

// func Test_Init(t *testing.T) {
// 	rh := RepoHandler{}
// 	rh.Init()
// 	server = httptest.NewServer(rh.router)
// 	testUrl = server.URL
// }

func TestAddRepos(t *testing.T) {

	ctrl := gomock.NewController(t)
	m := mockhelmclient.NewMockClient(ctrl)
	m.EXPECT().AddOrUpdateChartRepo(repo.Entry{
		Name: "test",
		URL:  "test-url",
	}).Return(nil)
	hcm := HelmClientManager{
		Hci: m,
	}

	defer ctrl.Finish()

	chartConfig := schemas.IndexFile{}
	byteChartConfig, _ := json.Marshal(chartConfig)

	repoConfig := schemas.Repository{
		Name: "test",
		Url:  "test-url",
	}
	byteRepoConfig, _ := json.Marshal(repoConfig)

	_ = os.WriteFile(repositoryConfig, byteRepoConfig, 0644)
	_ = os.WriteFile(repositoryCache+"/test"+indexFileSuffix, byteChartConfig, 0644)
	defer os.Remove(repositoryConfig)
	defer os.Remove(repositoryCache + "/test" + indexFileSuffix)

	t.Run("check add repos", func(t *testing.T) {
		repoReq := schemas.RepoRequest{
			Name:    "test",
			RepoURL: "test-url",
		}
		byteRepoReq, _ := json.Marshal(repoReq)
		reqBody := bytes.NewBuffer(byteRepoReq)
		req, err := http.NewRequest("POST", "/helm/repos", reqBody)
		assert.Nil(t, err, "")

		response := httptest.NewRecorder()
		hcm.AddChartRepo(response, req)
		if status := response.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})

}

// func TestGetRepos(t *testing.T) {

// 	ctrl := gomock.NewController(t)
// 	helmClient := mockhelmclient.NewMockClient(ctrl)
// 	hcm := HelmClientManager{
// 		Hci: helmClient,
// 	}
// 	defer ctrl.Finish()

// 	repoConfig := schemas.Repository{
// 		Name: "test",
// 		Url:  "test-url",
// 	}
// 	byteRepoConfig, _ := json.Marshal(repoConfig)

// 	_ = os.WriteFile(repositoryConfig, byteRepoConfig, 0644)
// 	defer os.Remove(repositoryConfig)

// 	t.Run("check get repos", func(t *testing.T) {

// 		req, err := http.NewRequest("GET", "/helm/repos", nil)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		response := httptest.NewRecorder()
// 		hcm.GetChartRepos(response, req)
// 		if status := response.Code; status != http.StatusOK {
// 			t.Errorf("handler returned wrong status code: got %v want %v",
// 				status, http.StatusOK)
// 		}
// 		content, _ := ioutil.ReadAll(response.Body)
// 		result := &schemas.RepoResponse{}
// 		_ = json.Unmarshal(content, result)

// 		for _, r := range result.RepoInfo {
// 			require.Equal(t, repoConfig.Name, r.Name)
// 		}

// 	})
// }
