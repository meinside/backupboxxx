# backupboxxx

Backups local files to your Dropbox account.

## Install

```bash
$ go install github.com/meinside/backupboxxx@latest
```

## Configuration

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

### Using Infisical

You can also use [Infisical](https://infisical.com/) for retrieving your access token:

```json
{
  "infisical": {
    "workspace_id": "012345abcdefg",
    "token": "st.xyzwabcd.0987654321.abcdefghijklmnop",
    "environment": "dev",
    "secret_type": "shared",
    "key_path": "/path/to/your/KEY_TO_ACCESS_TOKEN"
  }
}
```

If your Infisical workspace's E2EE setting is enabled, you also need to provide your API key:

```json
{
  "infisical": {
    "e2ee": true,
    "api_key": "ak.1234567890.abcdefghijk",

    "workspace_id": "012345abcdefg",
    "token": "st.xyzwabcd.0987654321.abcdefghijklmnop",
    "environment": "dev",
    "secret_type": "shared",
    "key_path": "/path/to/your/KEY_TO_ACCESS_TOKEN"
  }
}
```

## Backup List

After configuration, create a backup list file:

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

## Run

Now run with the backup list file:

```bash
$ backupboxxx [backup-list-filepath]
```

## License

MIT

