package atlas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// BuildConfig represents a Packer build configuration.
type BuildConfig struct {
	// User is the namespace under which the build config lives
	User string `json:"username"`

	// Name is the actual name of the build config, unique in the scope
	// of the username.
	Name string `json:"name"`
}

// BuildConfigVersion represents a single uploaded (or uploadable) version
// of a build configuration.
type BuildConfigVersion struct {
	// The fields below are the username/name combo to uniquely identify
	// a build config.
	User string `json:"username"`
	Name string `json:"name"`

	// Builds is the list of builds that this version supports.
	Builds []BuildConfigBuild
}

// BuildConfigBuild is a single build that is present in an uploaded
// build configuration.
type BuildConfigBuild struct {
	// Name is a unique name for this build
	Name string `json:"name"`

	// Type is the type of builder that this build needs to run on,
	// such as "amazon-ebs" or "qemu".
	Type string `json:"type"`
}

// BuildConfig gets a single build configuration by user and name.
func (c *Client) BuildConfig(user, name string) (*BuildConfig, error) {
	endpoint := fmt.Sprintf("/api/v1/packer/build-configurations/%s/%s", user, name)
	request, err := c.Request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	response, err := checkResp(c.HTTPClient.Do(request))
	if err != nil {
		return nil, err
	}

	var bc BuildConfig
	if err := decodeJSON(response, &bc); err != nil {
		return nil, err
	}

	return &bc, nil
}

// CreateBuildConfig creates a new build configuration.
func (c *Client) CreateBuildConfig(user, name string) error {
	endpoint := "/api/v1/packer/build-configurations"

	body, err := json.Marshal(&bcWrapper{
		BuildConfig: &BuildConfig{
			User: user,
			Name: name,
		},
	})
	if err != nil {
		return err
	}

	request, err := c.Request("POST", endpoint, &RequestOptions{
		Body: bytes.NewReader(body),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	})
	if err != nil {
		return err
	}

	_, err = checkResp(c.HTTPClient.Do(request))
	return err
}

// UploadBuildConfigVersion creates a single build configuration version
// and uploads the template associated with it.
//
// Actual API: "Create Build Config Version"
func (c *Client) UploadBuildConfigVersion(
	v *BuildConfigVersion, tpl io.Reader, size int64) error {
	endpoint := fmt.Sprintf("/api/v1/packer/build-configurations/%s/%s/versions",
		v.User, v.Name)

	var bodyData bcCreateWrapper
	bodyData.Version.Builds = v.Builds
	body, err := json.Marshal(bodyData)
	if err != nil {
		return err
	}

	request, err := c.Request("POST", endpoint, &RequestOptions{
		Body: bytes.NewReader(body),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	})
	if err != nil {
		return err
	}

	response, err := checkResp(c.HTTPClient.Do(request))
	if err != nil {
		return err
	}

	var data bcCreate
	if err := decodeJSON(response, &data); err != nil {
		return err
	}

	if err := c.putFile(data.UploadPath, tpl, size); err != nil {
		return err
	}

	return nil
}

// bcWrapper is the API wrapper since the server wraps the resulting object.
type bcWrapper struct {
	BuildConfig *BuildConfig `json:"build_configuration"`
}

// bcCreate is the struct returned when creating a build configuration.
type bcCreate struct {
	UploadPath string `json:"upload_path"`
}

// bcCreateWrapper is the wrapper for creating a build config.
type bcCreateWrapper struct {
	Version struct {
		Builds []BuildConfigBuild `json:"builds"`
	} `json:"version"`
}
