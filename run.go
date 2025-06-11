// run.go

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/fatih/color"
)

var _usersDir string

// initialize os-specific values
func init() {
	switch runtime.GOOS {
	case "darwin":
		_usersDir = "/Users"
	default:
		_usersDir = "/home"
	}
}

// upload file to Dropbox (ignore error)
func uploadFile(client files.Client, root string, path string, ignore []string) {
	if isInList(ignore, filepath.Base(path)) {
		printColored(color.FgYellow, "> ignoring: %s\n", path)
		return // skip
	}

	if stat, err := os.Stat(path); err == nil {
		if stat.IsDir() {
			if fs, err := os.ReadDir(path); err == nil {
				for _, file := range fs {
					uploadFile(client, root, filepath.Join(path, file.Name()), ignore)
				}
			} else {
				printColored(color.FgRed, "* error while recursing directory: %s (%s)\n", path, err)
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
					printColored(color.FgGreen, "> uploaded successfully: %s\n", path)
				} else {
					printColored(color.FgRed, "* error while uploading: %s (%s)\n", path, err)
				}
			} else {
				printColored(color.FgRed, "* error while reading file: %s (%s)\n", path, err)
			}
		}
	} else {
		printColored(color.FgRed, "* error while reading file: %s (%s)\n", path, err)
	}
}

// do backup with given backup file list
func backup(
	client files.Client,
	binPath string,
	backupListFilepath string,
) error {
	printColored(color.FgHiWhite, "> reading backup list file: %s\n", backupListFilepath)

	list, err := readBackupList(backupListFilepath)
	if err != nil {
		return fmt.Errorf("backup failed: %s", err)
	}

	dirname := list.Dirname
	fs := list.Files
	ignore := list.Ignore

	printColored(color.FgHiWhite, "> destination dir: %s\n", dirname)

	for _, file := range fs {
		uploadFile(client, dirname, expandPath(file, binPath), ignore)
	}

	return nil
}

// read backup list file
func readBackupList(path string) (list *backupList, err error) {
	list = new(backupList)
	if _, err = os.Stat(path); err == nil {
		var bytes []byte
		if bytes, err = os.ReadFile(path); err == nil {
			if bytes, err = standardizeJSON(bytes); err == nil {
				if err = json.Unmarshal(bytes, &list); err == nil {
					return list, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("failed to read backup list: %s", err)
}

// expand given path
func expandPath(path, binPath string) string {
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
		if dir, err := filepath.Abs(filepath.Dir(binPath)); err == nil {
			path = filepath.Join(dir, path)
		}
	}
	return path
}

// check if given element is in the list or not
func isInList(list []string, element string) bool {
	return slices.Contains(list, element)
}

// print usage and exit(0)
func printUsage(binPath string) {
	printColored(
		color.FgWhite,
		`
> usage:

# show this message
$ %[1]v -h
$ %[1]v --help

# print out a sample backup list file
$ %[1]v -g
$ %[1]v --generate

# do backup
$ %[1]v backup_list.json

`,
		filepath.Base(binPath),
	)

	os.Exit(0)
}

// print sample list and exit(0)
func printSampleList() {
	printColored(
		color.FgCyan,
		`
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
`,
	)

	os.Exit(0)
}

// run application
func run(args []string) {
	var conf config
	var err error

	// help
	if isInList(args, "-h") || isInList(args, "--help") || len(args) < 2 {
		printUsage(args[0])
	}

	// generate a list file
	if isInList(args, "-g") || isInList(args, "--generate") {
		printSampleList()
	}

	// load configuration and do backup
	if conf, err = loadConf(); err == nil {
		var token *string
		if token, err = conf.getAccessToken(); err == nil {
			err = backup(
				files.New(dropbox.Config{Token: *token}),
				args[0],
				args[1],
			)

			// success
			if err == nil {
				os.Exit(0)
			}
		}
	}

	// exit with error
	printErrorAndExit(err)
}
