package repos

// import (
// 	"bytes"
// 	"crypto/tls"
// 	"encoding/json"
// 	"net/http"

// 	"github.com/tmax-cloud/helm-apiserver/pkg/schemas"
// 	"helm.sh/helm/v3/pkg/repo"

// 	"k8s.io/klog"
// )

// // [TODO] : channel 사용으로 sync call 되도록 변경
// func (hcm *HelmClientManager) CreateChartRepo(w http.ResponseWriter, r *http.Request) {
// 	klog.Infoln("Create Helm Chart Repository")

// 	req := schemas.RepoClientRequest{}
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		klog.Errorln(err, "failed to decode request")
// 		respond(w, http.StatusBadRequest, &schemas.Error{
// 			Error:       err.Error(),
// 			Description: "Error occurs while decoding request",
// 		})
// 		return
// 	}

// 	createReq := schemas.RepoClientRequest{
// 		Name:      req.Name,
// 		Auto_init: true,
// 	}
// 	byteReq, _ := json.Marshal(createReq)
// 	reqBody := bytes.NewBuffer(byteReq)
// 	postReq, _ := http.NewRequest("POST", "https://api.github.com/user/repos", reqBody)
// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			TLSClientConfig: &tls.Config{
// 				InsecureSkipVerify: true,
// 			},
// 		},
// 	}
// 	postReq.Header.Add("Content-Type", "application/json")
// 	postReq.Header.Add("Accept", "application/vnd.github.v3+json")
// 	postReq.Header.Add("Authorization", "token"+" "+"Access token") // Personal Access Token 필요
// 	postResp, err := client.Do(postReq)
// 	if err != nil { // 에러 체크는 response statuscode로 하기
// 		klog.Errorln(err)
// 		return
// 	}
// 	klog.Info(postResp.StatusCode)
// 	defer postResp.Body.Close()

// 	PageReq := &schemas.GithubPageRequest{
// 		Source: schemas.Source{
// 			Branch: "main",
// 			Path:   "/",
// 		},
// 	}
// 	bytePageReq, _ := json.Marshal(PageReq)
// 	pageReqBody := bytes.NewBuffer(bytePageReq)
// 	pageReq, _ := http.NewRequest("POST", "https://api.github.com/repos/min-charles/gh-test/pages", pageReqBody)
// 	// client := &http.Client{
// 	// 	Transport: &http.Transport{
// 	// 		TLSClientConfig: &tls.Config{
// 	// 			InsecureSkipVerify: true,
// 	// 		},
// 	// 	},
// 	// }
// 	pageReq.Header.Add("Content-Type", "application/json")
// 	pageReq.Header.Add("Accept", "application/vnd.github.v3+json")
// 	pageReq.Header.Add("Authorization", "token"+" "+"Access token") // Personal Access Token 필요
// 	pageResp, err := client.Do(pageReq)
// 	if err != nil {
// 		klog.Errorln(err)
// 		return
// 	}
// 	klog.Info(pageResp.StatusCode)
// 	klog.Info(*pageResp)
// 	defer pageResp.Body.Close()

// 	chartRepo := repo.Entry{
// 		Name: req.Name,
// 		URL:  "https://api.github.com/repos/min-charles/gh-test/pages",
// 	}

// 	// github의 경우 index.yaml이 있어야 repo 추가 가능
// 	if err := hcm.Hci.AddOrUpdateChartRepo(chartRepo); err != nil {
// 		klog.Errorln(err, "failed to add chart repo")
// 		respond(w, http.StatusBadRequest, &schemas.Error{
// 			Error:       err.Error(),
// 			Description: "Error occurs while adding helm repo",
// 		})
// 		return
// 	}

// 	respond(w, http.StatusOK, "test done")
// }

// func (hcm *HelmClientManager) CreateChartRepo(w http.ResponseWriter, r *http.Request) {
// 	klog.Infoln("Create Helm Chart Repository")

// 	req := schemas.RepoClientRequest{}
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		klog.Errorln(err, "failed to decode request")
// 		respond(w, http.StatusBadRequest, &schemas.Error{
// 			Error:       err.Error(),
// 			Description: "Error occurs while decoding request",
// 		})
// 		return
// 	}

// 	byteReq, _ := json.Marshal(req)
// 	reqBody := bytes.NewBuffer(byteReq)
// 	postReq, _ := http.NewRequest("POST", "https://gitlab.gitlab-system.172.22.11.16.nip.io/api/v4/projects", reqBody)
// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			TLSClientConfig: &tls.Config{
// 				InsecureSkipVerify: true,
// 			},
// 		},
// 	}
// 	postReq.Header.Add("Content-Type", "application/json")
// 	postReq.Header.Add("PRIVATE-TOKEN", "Token")
// 	postResp, err := client.Do(postReq)
// 	if err != nil {
// 		klog.Errorln(err)
// 		return
// 	}
// 	defer postResp.Body.Close()

// 	getReq, _ := http.NewRequest("GET", "https://gitlab.gitlab-system.172.22.11.16.nip.io/api/v4/projects?search="+req.Name, nil)
// 	getReq.Header.Add("PRIVATE-TOKEN", "Token")
// 	getResp, err := client.Do(getReq)
// 	if err != nil {
// 		klog.Errorln(err)
// 		return
// 	}

// 	getRespBody, _ := ioutil.ReadAll(getResp.Body)
// 	defer getResp.Body.Close()

// 	var parser []interface{}
// 	err = json.Unmarshal(getRespBody, &parser)
// 	if err != nil {
// 		klog.Errorln(err)
// 	}

// 	projectId := fmt.Sprintf("%.0f", parser[0].(map[string]interface{})["id"]) // get project ID for add chart repo
// 	klog.Info(projectId)

// 	if getResp.StatusCode >= 400 {
// 		klog.Info("Create Repo is failed." + getResp.Status)
// 		return
// 	}

// 	chartRepo := repo.Entry{
// 		Name:                  req.Name,
// 		URL:                   "https://gitlab.gitlab-system.172.22.11.16.nip.io/api/v4/projects/" + projectId + "/packages/helm/stable",
// 		Username:              "root",
// 		Password:              "Token",
// 		CAFile:                ca_crt,
// 		CertFile:              public_key,
// 		KeyFile:               private_key,
// 		InsecureSkipTLSverify: true,
// 	}

// 	// hcm.SetClientTLS(chartRepo.URL)

// 	if err := hcm.Hci.AddOrUpdateChartRepo(chartRepo); err != nil {
// 		klog.Errorln(err, "failed to add chart repo")
// 		respond(w, http.StatusBadRequest, &schemas.Error{
// 			Error:       err.Error(),
// 			Description: "Error occurs while adding helm repo",
// 		})
// 		return
// 	}

// 	respond(w, http.StatusOK, "test done")
// }
