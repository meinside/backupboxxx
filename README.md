# backupboxxx

Backups local files to your Dropbox account.

## install

```bash
$ go get -u github.com/meinside/backupboxxx
```

## setup

Get your access token from:

```
Developers page > App console > [Your App] > Settings > OAuth2 > Generated access token > Generate
```

then create a file named `backupboxxx.json` in your `$HOME/.config/` directory:

```json
{
	"access_token": "PUT_YOUR_GENERATED_ACCESS_TOKEN_HERE"
}
```

After that, create a backup list file:

```json
{
	"dirname": "backup_20190605",
	"files": [
		"~/.zshrc",
		"~/files/photos",
		"~someusername/somewhere/filename",
		"/etc/hosts"
	],
	"ignore": [
		".ssh",
		".git",
		".svn",
		".DS_Store"
	]
}
```

## run

```bash
$ $GOPATH/bin/backupboxxx [backup-list-filepath]
```

