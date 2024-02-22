package main

import (
	"encoding/json"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"

	"github.com/tailscale/hujson"

	"github.com/meinside/infisical-go"
	"github.com/meinside/infisical-go/helper"
)

const (
	applicationName = "backupboxxx"
	configFilename  = "config.json"
)

type config struct {
	// Developers page > App console > [Your App] > Settings > OAuth2 > Generated access token > Generate
	//
	// example:
	//
	// {
	//   "access_token": "abcdefghijklmnopqrstuvwxyz0123456789"
	// }
	AccessToken string `json:"access_token,omitempty"` // Dropbox access token

	// or Infisical settings
	Infisical *struct {
		// for universal-auth of Infisical
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`

		WorkspaceID        string               `json:"workspace_id"`
		Environment        string               `json:"environment"`
		SecretType         infisical.SecretType `json:"secret_type"`
		AccessTokenKeyPath string               `json:"key_path"`
	} `json:"infisical,omitempty"`
}

// get access token, retrieve it from infisical if needed
func (c *config) GetAccessToken() string {
	if c.AccessToken == "" && c.Infisical != nil {
		var accessToken string

		// read access token from infisical
		var err error
		accessToken, err = helper.Value(
			c.Infisical.ClientID,
			c.Infisical.ClientSecret,
			c.Infisical.WorkspaceID,
			c.Infisical.Environment,
			c.Infisical.SecretType,
			c.Infisical.AccessTokenKeyPath,
		)

		if err != nil {
			_stderr.Printf("* failed to retrieve Dropbox access token from infisical: %s\n", err)
		}

		c.AccessToken = accessToken
	}

	return c.AccessToken
}

var _usersDir string

// loggers
var _stdout = log.New(os.Stdout, "", 0)
var _stderr = log.New(os.Stderr, "", 0)

// setup os-specific values
func init() {
	switch runtime.GOOS {
	case "darwin":
		_usersDir = "/Users"
	default:
		_usersDir = "/home"
	}
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
			_stderr.Fatalf("* failed to get home directory (%s)\n", err)
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

// BackupList for listing files to backup
type BackupList struct {
	Dirname string   `json:"dirname"`
	Files   []string `json:"files"`
	Ignore  []string `json:"ignore"`
}

// upload file to Dropbox
func uploadFile(client files.Client, root string, path string, ignore []string) {
	if isInList(ignore, filepath.Base(path)) {
		_stdout.Printf("> ignoring: %s\n", path)
		return //skip
	}

	if stat, err := os.Stat(path); err == nil {
		if stat.IsDir() {
			if _files, err := os.ReadDir(path); err == nil {
				for _, file := range _files {
					uploadFile(client, root, filepath.Join(path, file.Name()), ignore)
				}
			} else {
				_stderr.Printf("* error while recursing directory: %s (%s)\n", path, err)
			}
		} else {
			if reader, err := os.Open(path); err == nil {
				defer reader.Close()

				if _, err := client.Upload(&files.UploadArg{
					CommitInfo: files.CommitInfo{
						Path:       filepath.Join("/", root, path),
						Mode:       &files.WriteMode{Tagged: dropbox.Tagged{Tag: "overwrite"}}, // overwrite!
						Autorename: false,
						Mute:       false,
					},
				}, reader); err == nil {
					_stdout.Printf("> uploaded successfully: %s\n", path)
				} else {
					_stderr.Printf("* error while uploading: %s (%s)\n", path, err)
				}
			} else {
				_stderr.Printf("* error while reading file: %s (%s)\n", path, err)
			}
		}
	} else {
		_stderr.Printf("* error while reading file: %s (%s)\n", path, err)
	}
}

// do backup with given backup file list
func backup(client files.Client, backupListFilepath string) {
	list := readBackupList(backupListFilepath)

	dirname := list.Dirname
	_files := list.Files
	ignore := list.Ignore

	_stdout.Printf("> destination dir: %s\n", dirname)

	for _, file := range _files {
		uploadFile(client, dirname, expandPath(file), ignore)
	}
}

// read backup list file
func readBackupList(path string) *BackupList {
	_stdout.Printf("> reading backup list file: %s\n", path)

	list := new(BackupList)
	if _, err := os.Stat(path); err != nil {
		_stderr.Fatalf("* failed to stat backup list file (%s)\n", err)
	} else {
		if bytes, err := os.ReadFile(path); err != nil {
			_stderr.Fatalf("* failed to read backup list file (%s)\n", err)
		} else {
			if bytes, err := standardizeJSON(bytes); err == nil {
				if err := json.Unmarshal(bytes, &list); err != nil {
					_stderr.Fatalf("* failed to parse backup list file (%s)\n", err)
				}
			}
		}
	}
	return list
}

// expand given path
func expandPath(path string) string {
	pathSeparator := string(filepath.Separator)

	if strings.HasPrefix(path, pathSeparator) { // case 1: /some/absolute/path
		// do nothing
	} else if strings.HasPrefix(path, "~"+pathSeparator) { // case 2: ~/somewhere
		// replace "~/" with user's home path
		if currentUser, err := user.Current(); err == nil {
			path = strings.Replace(path, "~", currentUser.HomeDir, 1)
		}
	} else if strings.HasPrefix(path, "~") { // case 3: ~someone/somewhere
		// replace "~" with "/home/" or "/Users/"
		path = strings.Replace(path, "~", _usersDir+pathSeparator, 1)
	} else { // case 4: some/relative/path
		// prepend current directory
		if dir, err := filepath.Abs(filepath.Dir(os.Args[0])); err == nil {
			path = filepath.Join(dir, path)
		}
	}
	return path
}

// check if given element is in the list or not
func isInList(list []string, element string) bool {
	for _, value := range list {
		if value == element {
			return true
		}
	}
	return false
}

// print usage
func printUsage() {
	_stdout.Printf(`
> usage:

# show this message
$ %[1]v -h
$ %[1]v --help

# print out a sample backup list file
$ %[1]v -g
$ %[1]v --generate

# do backup
$ %[1]v backup_list.json

`, filepath.Base(os.Args[0]))
}

// print sample list
func printSampleList() {
	_stdout.Printf(`
// sample list in JSON(JWCC)
{
	// destination directory's name
	"dirname": "backup_20190605",

	// file paths that will be backed up
	"files": [
		"/etc/sysctl.conf",
		"/etc/dhcp/dhclient.conf",
		"/etc/samba/smb.conf",
		"~/.custom_aliases",
		"~/files/photos",
	],

	// names that will be ignored
	"ignore": [
		".ssh",
		".git",
		".svn",
		".DS_Store",
	],
}
`)
}

// print error and exit
func printErrorAndExit(err error) {
	_stderr.Fatalf(err.Error())
}

func main() {
	var conf config
	var err error

	// help
	if isInList(os.Args, "-h") || isInList(os.Args, "--help") || len(os.Args) < 2 {
		printUsage()

		os.Exit(0)
	}

	// generate a list file
	if isInList(os.Args, "-g") || isInList(os.Args, "--generate") {
		printSampleList()

		os.Exit(0)
	}

	// load configuration and do backup
	if conf, err = loadConf(); err == nil {
		backup(files.New(dropbox.Config{Token: conf.GetAccessToken()}), os.Args[1])
	} else {
		printErrorAndExit(err)
	}
}
