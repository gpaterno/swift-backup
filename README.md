# swift-backup

swift-backup is a simple one-binary program to backup/archive a single file on an host.
This was initially done to backup the mysql dump of the database.

```
Usage of swift-backup:
  -delete-after
    	Delete source file afterwards
  -flushlog string
    	sets the flush trigger level (default "none")
  -log string
    	sets the logging threshold (default "info")
  -os-authurl string
    	Keystone endpoint (default "https://os.ch-ti1.server.one/v2.0/")
  -os-password string
    	Swift password (default "password")
  -os-project string
    	Project name (default "admin")
  -os-region string
    	Region name (default "ch-ti1")
  -os-username string
    	Swift username (default "admin")
  -stderr
    	outputs to standard error (stderr)
```
