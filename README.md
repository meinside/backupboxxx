# backupboxxx

Backups local files to your Dropbox account.

## install

```bash
$ go install github.com/meinside/backupboxxx@latest
```

## setup

Get your Dropbox access token from:

```
Developers page > App console > [Your App] > Settings > OAuth2 > Generated access token > Generate
```

then create a file named `config.json` in `$XDG_CONFIG_HOME/backupboxxx/` or `$HOME/.config/backupboxxx` directory:

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

Now run with the backup list file:

```bash
$ $GOPATH/bin/backupboxxx [backup-list-filepath]
```

## license

MIT

