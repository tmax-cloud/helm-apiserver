package schemas

import (
	"helm.sh/helm/v3/pkg/time"
)

type RepoRequest struct {
	Name    string `json:"name,omitempty"`
	RepoURL string `json:"repoURL,omitempty"`
}

type RepoResponse struct {
	RepoInfo []Repository `json:"repoInfo,omitempty"`
}

type RepositoryFile struct {
	ApiVersion   string       `json:"apiVersion,omitempty"`
	Generated    time.Time    `json:"generated,omitempty"`
	Repositories []Repository `json:"repositories,omitempty"`
}

type Repository struct {
	CaFile                   string `json:"caFile,omitempty"`
	CertFile                 string `json:"certFile,omitempty"`
	Insecure_skip_tls_verify bool   `json:"insecure_skip_tls_verify,omitempty"`
	KeyFile                  string `json:"keyFile,omitempty"`
	Name                     string `json:"name,omitempty"`
	Pass_credentials_all     bool   `json:"pass_credentials_all,omitempty"`
	Password                 string `json:"password,omitempty"`
	Url                      string `json:"url,omitempty"`
	UserName                 string `json:"username,omitempty"`
}

type RepoClientRequest struct {
	Name      string `json:"name,omitempty"`
	Auto_init bool   `json:"auto_init,omitempty"`
}

type GithubPageRequest struct {
	Source Source `json:"source,omitempty"`
}

type Source struct {
	Branch string `json:"branch,omitempty"`
	Path   string `json:"path,omitempty"`
}
