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

func ReadRepoIndex(repoName string) (index *schemas.IndexFile, err error) {
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
