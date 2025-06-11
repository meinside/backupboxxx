// config.go

package main

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"path/filepath"

	"github.com/fatih/color"
	infisical "github.com/infisical/go-sdk"
	"github.com/infisical/go-sdk/packages/models"
	"github.com/tailscale/hujson"
)

const (
	applicationName = "backupboxxx"
	configFilename  = "config.json"
)

// config struct
type config struct {
	// Developers page > App console > [Your App] > Settings > OAuth2 > Generated access token > Generate
	//
	// example:
	//
	// {
	//   "access_token": "abcdefghijklmnopqrstuvwxyz0123456789"
	// }
	AccessToken *string `json:"access_token,omitempty"` // Dropbox access token

	// or Infisical settings
	Infisical *struct {
		// for universal-auth of Infisical
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`

		ProjectID          string `json:"project_id"`
		Environment        string `json:"environment"`
		SecretType         string `json:"secret_type"`
		AccessTokenKeyPath string `json:"key_path"`
	} `json:"infisical,omitempty"`
}

// return your access token of Dropbox
//
// (retrieve it from infisical if needed)
func (c *config) getAccessToken() (accessToken *string, err error) {
	if (c.AccessToken == nil || len(*c.AccessToken) == 0) &&
		c.Infisical != nil {
		// read access token from infisical
		client := infisical.NewInfisicalClient(context.TODO(), infisical.Config{
			SiteUrl: "https://app.infisical.com",
		})

		_, err = client.Auth().UniversalAuthLogin(c.Infisical.ClientID, c.Infisical.ClientSecret)
		if err != nil {
			printColored(color.FgHiRed, "* failed to authenticate with Infisical: %s", err)
			return nil, err
		}

		var secret models.Secret
		secret, err = client.Secrets().Retrieve(infisical.RetrieveSecretOptions{
			SecretKey:   path.Base(c.Infisical.AccessTokenKeyPath),
			SecretPath:  path.Dir(c.Infisical.AccessTokenKeyPath),
			ProjectID:   c.Infisical.ProjectID,
			Type:        c.Infisical.SecretType,
			Environment: c.Infisical.Environment,
		})
		if err != nil {
			printColored(color.FgHiRed, "* failed to retrieve Dropbox access token from infisical: %s\n", err)
			return nil, err
		}

		c.AccessToken = &secret.SecretValue
	}

	return c.AccessToken, nil
}

// standardize given JSON (JWCC) bytes
func standardizeJSON(b []byte) ([]byte, error) {
	ast, err := hujson.Parse(b)
	if err != nil {
		return b, err
	}
	ast.Standardize()

	return ast.Pack(), nil
}

// load config file
func loadConf() (conf config, err error) {
	// https://xdgbasedirectoryspecification.com
	configDir := os.Getenv("XDG_CONFIG_HOME")

	// If the value of the environment variable is unset, empty, or not an absolute path, use the default
	if configDir == "" || configDir[0:1] != "/" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			printColored(color.FgHiRed, "* failed to get home directory (%s)\n", err)
		} else {
			configDir = filepath.Join(homeDir, ".config", applicationName)
		}
	} else {
		configDir = filepath.Join(configDir, applicationName)
	}

	configFilepath := filepath.Join(configDir, configFilename)

	var bytes []byte
	if bytes, err = os.ReadFile(configFilepath); err == nil {
		if bytes, err = standardizeJSON(bytes); err == nil {
			if err = json.Unmarshal(bytes, &conf); err == nil {
				return conf, err
			}
		}
	}

	return config{}, err
}

// backupList for listing files to backup
type backupList struct {
	Dirname string   `json:"dirname"`
	Files   []string `json:"files"`
	Ignore  []string `json:"ignore"`
}
