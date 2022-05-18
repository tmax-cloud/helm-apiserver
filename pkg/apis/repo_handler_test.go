package apis

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"helm.sh/helm/v3/pkg/repo"

	// helmclient "github.com/mittwald/go-helm-client"
	gomock "github.com/golang/mock/gomock"
	mockhelmclient "github.com/mittwald/go-helm-client/mock"
	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
	// "github.com/tmax-cloud/helm-apiserver/internal"
)

func TestAddRepos(t *testing.T) {

	ctrl := gomock.NewController(t)
	m := mockhelmclient.NewMockClient(ctrl)
	m.EXPECT().AddOrUpdateChartRepo(repo.Entry{
		Name:                  "test",
		URL:                   "test-url",
		InsecureSkipTLSverify: true, // for-test
	}).Return(nil)
	hcm := HelmClientManager{
		Hci: m,
	}

	defer ctrl.Finish()

	chartConfig := schemas.IndexFile{}
	byteChartConfig, _ := json.Marshal(chartConfig)

	repoConfig := schemas.RepositoryFile{}
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

func TestDeleteRepos(t *testing.T) {

	ctrl := gomock.NewController(t)
	m := mockhelmclient.NewMockClient(ctrl)
	hcm := &HelmClientManager{
		Hci: m,
	}

	defer ctrl.Finish()

	chartConfig := schemas.IndexFile{
		APIVersion: "test",
	}

	byteChartConfig, _ := json.Marshal(chartConfig)

	repoConfig := schemas.RepositoryFile{}
	testRepo := schemas.Repository{
		Name: "test",
		Url:  "test-url",
	}
	repoConfig.Repositories = append(repoConfig.Repositories, testRepo)
	byteRepoConfig, _ := json.Marshal(repoConfig)

	os.Mkdir(repositoryCache, 0755)
	os.WriteFile(repositoryConfig, byteRepoConfig, 0644)
	os.WriteFile(repositoryCache+"/test"+indexFileSuffix, byteChartConfig, 0644)
	os.WriteFile(repositoryCache+"/test"+chartsFileSuffix, byteChartConfig, 0644)

	defer os.Remove(repositoryConfig)
	defer os.Remove(repositoryCache + "/test" + indexFileSuffix)
	defer os.Remove(repositoryCache + "/test" + chartsFileSuffix)
	defer os.Remove(repositoryCache)

	t.Run("check delete repos", func(t *testing.T) {

		req := httptest.NewRequest("DELETE", "/helm/repos/test", nil)
		req = mux.SetURLVars(req, map[string]string{"repo-name": "test"})

		response := httptest.NewRecorder()
		hcm.DeleteChartRepo(response, req)
		if status := response.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})

	afterByteRepoConfig, _ := os.ReadFile(repositoryConfig)
	afterRepoConfig := &schemas.RepositoryFile{}
	json.Unmarshal(afterByteRepoConfig, afterRepoConfig)

	for _, r := range afterRepoConfig.Repositories {
		if r.Name == "test" {
			t.Error("test repo is not deleted")
		}
	}

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
