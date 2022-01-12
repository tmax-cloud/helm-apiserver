package schemas

type ReleaseRequest struct {
	Name      string             `json:"name"`
	Namespace string             `json:"namespace"`
	Spec      ReleaseRequestSpec `json:"spec,omitempty"`
	Values    string             `json:"values,omitempty"`
}

type ReleaseRequestSpec struct {
	Repository  string `json:"repository"`
	Path        string `json:"path"`
	ReleaseName string `json:"releasename"`
	Version     string `json:"version"`
}
