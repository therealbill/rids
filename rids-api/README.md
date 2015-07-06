# Redis Instance Data Service

A small HTTPS enabled API for uploading data about Redis instances. This
includes such things as connection information, including auth, command stats,
and anything else you want to store.

# Building

Standard Go build process.


# Configuring and Running

All configuration is done via command line flags or via environment variables
(or a mix as desired). Run with '-help' to see the details.

Included are systemd unit files with sensible defaults. 

 * Place the service file in /lib/systemd/system/
 * Create the directory /lib/systemd/system/rids.d
 * Place the environment.conf in that directory
 * Modify the environment file as needed
 * execute `systemctl daemon-reload`
 * Start the service (`systemctl start rids`)

## SSL 
You will need certificate and key files. BY default they are expected in :
* `/etc/rids/cert.pem`
* `/etc/rids/key.pem`

The full path to the file can be provided via CLI flags as well as the
environment variables `CERTFILE` and `KEYFILE`.

