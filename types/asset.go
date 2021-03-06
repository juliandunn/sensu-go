package types

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"path"
	"regexp"

	"github.com/sensu/sensu-go/util/eval"
)

// AssetNameRegexStr used to validate name of asset
var AssetNameRegexStr = `[a-z0-9\/\_\.\-]+`

// AssetNameRegex used to validate name of asset
var AssetNameRegex = regexp.MustCompile("^" + AssetNameRegexStr + "$")

// Validate returns an error if the asset contains invalid values.
func (a *Asset) Validate() error {
	if err := ValidateAssetName(a.Name); err != nil {
		return err
	}

	if a.Organization == "" {
		return errors.New("organization cannot be empty")
	}

	if a.Sha512 == "" {
		return errors.New("SHA-512 checksum cannot be empty")
	}

	if a.URL == "" {
		return errors.New("URL cannot be empty")
	}

	u, err := url.Parse(a.URL)
	if err != nil {
		return errors.New("invalid URL provided")
	}

	if u.Scheme != "https" && u.Scheme != "http" {
		return errors.New("URL must be HTTP or HTTPS")
	}

	return eval.ValidateStatements(a.Filters)
}

// GetEnvironment refers to the organization the check belongs to
func (a *Asset) GetEnvironment() string {
	return ""
}

// ValidateAssetName validates that asset's name is valid
func ValidateAssetName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	if !AssetNameRegex.MatchString(name) {
		return errors.New(
			"name must be lowercase and may only contain forward slashes, underscores, dashes and numbers",
		)
	}

	return nil
}

// Filename returns the filename of the underlying asset; pulled from the URL
func (a *Asset) Filename() string {
	u, err := url.Parse(a.URL)
	if err != nil {
		return ""
	}

	_, file := path.Split(u.EscapedPath())
	return file
}

// FixtureAsset given a name returns a valid asset for use in tests
func FixtureAsset(name string) *Asset {
	bytes := make([]byte, 10)
	_, _ = rand.Read(bytes)
	hash := hex.EncodeToString(bytes)

	return &Asset{
		Name:   name,
		Sha512: "25e01b962045f4f5b624c3e47e782bef65c6c82602524dc569a8431b76cc1f57639d267380a7ec49f70876339ae261704fc51ed2fc520513cf94bc45ed7f6e17",
		URL:    "https://localhost/" + hash + ".zip",
		Metadata: map[string]string{
			"Content-Type":            "application/zip",
			"X-Intended-Distribution": "trusty-14",
		},
		Organization: "default",
	}
}
