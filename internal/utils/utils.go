package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/ghodss/yaml"
	gsocket "github.com/gorilla/websocket"
	"k8s.io/klog"

	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
)

const (
	indexFileSuffix  = "-index.yaml"
	chartsFileSuffix = "-charts.txt"

	repositoryCache  = "/tmp/.helmcache" // 캐시 디렉토리. 특정 chart-repo에 해당하는 chart 이름 리스트 txt파일과, 해당 repo의 index.yaml 파일이 저장됨
	repositoryConfig = "/tmp/.helmrepo"  // 현재 add된 repo들 저장.

	defaultTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

// FileExists checks if the file exists in path
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func UpgradeWebsocket(res http.ResponseWriter, req *http.Request) (*gsocket.Conn, error) {
	upgrader := gsocket.Upgrader{
		// TODO : FIX ME for specific domain
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	c, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		klog.Errorln(err)
		return nil, err
	}
	return c, err
}

// RespondJSON responds with arbitrary data objects
func RespondJSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")

	j, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = w.Write(j)
	if err != nil {
		return err
	}

	return nil
}

func Respond(w http.ResponseWriter, statusCode int, body interface{}) {
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		klog.Errorln(err, "Error occurs while encoding response body")
	}
}

// CORS 에러로 인해 header 추가 21.04.19
func SetResponseHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, OPTIONS") // Method 각각 써줘야 할것 같음
	w.Header().Set("Access-Control-Allow-Headers", "X-CSRF-Token, X-Requested-With, Accept, Accept-Version, Content-Length, Content-MD5, Content-Type, Date, X-Api-Version")

}

// ErrorResponse is a common struct for responding error for HTTP requests
type ErrorResponse struct {
	Message string `json:"message"`
}

// RespondError responds to a HTTP request with body of ErrorResponse
func RespondError(w http.ResponseWriter, code int, msg string) error {
	w.WriteHeader(code)
	return RespondJSON(w, ErrorResponse{Message: msg})
}

func GetIndex() *schemas.IndexFile {
	// Read repositoryConfig File which contains repo Info list
	repoList, err := ReadRepoList()
	if err != nil {
		klog.Errorln(err, "failed to save index file")
		return nil
	}

	repoInfos := make(map[string]string)
	// store repo names
	if len(repoList.Repositories) != 0 {
		for _, repoInfo := range repoList.Repositories {
			repoInfos[repoInfo.Name] = repoInfo.Url
		}
	}

	index := &schemas.IndexFile{}
	allEntries := make(map[string]schemas.ChartVersions)

	// read all index.yaml file and save only Entries
	for repoName, repoUrl := range repoInfos {
		if index, err = ReadRepoIndex(repoName); err != nil {
			klog.Errorln(err, "failed to read index file")
		}

		// add repo info
		for key, charts := range index.Entries {
			for _, chart := range charts {
				chart.Repo.Name = repoName
				chart.Repo.Url = repoUrl
			}
			allEntries[repoName+"_"+key] = charts // 중복 chart name 가능하도록 repo name과 결합
		}
	}

	index.Entries = allEntries
	klog.V(5).Info("saving index file is done")
	return index
}

func GetSingleChart(index *schemas.IndexFile) map[string]schemas.ChartVersions {
	if index == nil {
		return nil
	}

	singleChartEntries := make(map[string]schemas.ChartVersions)
	for key, charts := range index.Entries {
		var oneChart []*schemas.ChartVersion
		oneChart = append(oneChart, charts[0])
		singleChartEntries[key] = oneChart
	}
	return singleChartEntries
}

func ReadRepoIndex(repoName string) (index *schemas.IndexFile, err error) {
	if _, err := os.Stat(repositoryCache + "/" + repoName + indexFileSuffix); errors.Is(err, os.ErrNotExist) {
		klog.Info(repoName + "-index.yaml file is not exist")
		return nil, err
	}

	index = &schemas.IndexFile{}
	indexFile, err := ioutil.ReadFile(repositoryCache + "/" + repoName + indexFileSuffix)
	if err != nil {
		klog.Errorln(err, "failed to read index.yaml file of "+repoName)
		return nil, err
	}

	indexFileJson, _ := yaml.YAMLToJSON(indexFile) // Should transform yaml to Json

	if err := json.Unmarshal(indexFileJson, index); err != nil {
		klog.Errorln(err, "failed to unmarshal index file")
		return nil, err
	}

	return index, nil
}

func ReadRepoList() (repoList *schemas.RepositoryFile, err error) {
	if _, err := os.Stat(repositoryConfig); errors.Is(err, os.ErrNotExist) {
		klog.Info("No Helm chart repository is added")
		return nil, err
	}

	repoList = &schemas.RepositoryFile{}
	repoListFile, err := ioutil.ReadFile(repositoryConfig)
	if err != nil {
		klog.Errorln(err, "failed to get repository list")
		return nil, err
	}

	repoListFileJson, _ := yaml.YAMLToJSON(repoListFile) // Should transform yaml to Json

	if err = json.Unmarshal(repoListFileJson, repoList); err != nil {
		klog.Errorln(err, "failed to unmarshal repo file")
		return nil, err
	}

	return repoList, nil
}

func ReadDefaultToken() (string, error) {
	defaultToken, err := ioutil.ReadFile(defaultTokenPath)
	if err != nil {
		return "", err
	}

	return string(defaultToken), nil
}
