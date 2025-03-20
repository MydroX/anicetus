package models

type Browser struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	Engine        string `json:"engine"`
	EngineVersion string `json:"engine_version"`
}

type OS struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Architecture string `json:"architecture"`
}

type Device struct {
	Name string `json:"name"`
}
