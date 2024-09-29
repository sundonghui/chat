package model

// VersionInfo Model
//
// swagger:model VersionInfo
type VersionInfo struct {
	// The current version.
	//
	// required: true
	// example: 0.0.1
	Version string `json:"version"`
	// The git commit hash on which this binary was built.
	//
	// required: true
	// example: 3fe4992a28e4c224b5fb071b0943b53a1814a8bb
	Commit string `json:"commit"`
	// The date on which this binary was built.
	//
	// required: true
	// example: 2018-02-27T19:36:10.5045044+01:00
	BuildDate string `json:"buildDate"`
}
