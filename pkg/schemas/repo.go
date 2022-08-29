package schemas

import (
	"helm.sh/helm/v3/pkg/time"
)

type RepoRequest struct {
	Name       string `json:"name,omitempty"`
	RepoURL    string `json:"repoURL,omitempty"`
	Is_private bool   `json:"is_private,omitempty"`
	Id         string `json:"id,omitempty"`
	Password   string `json:"password,omitempty"`
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
	CaFile                   string    `json:"-"`
	CertFile                 string    `json:"-"`
	Insecure_skip_tls_verify bool      `json:"-"`
	KeyFile                  string    `json:"-"`
	Name                     string    `json:"name,omitempty"`
	Pass_credentials_all     bool      `json:"-"`
	Password                 string    `json:"-"`
	Url                      string    `json:"url,omitempty"`
	UserName                 string    `json:"-"`
	LastUpdated              time.Time `json:"lastupdated,omitempty"`
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
