# CD Watcher
A single binary for performing the CD part of a CI/CD pipeline

## Dependencies

This depends on systemd credentials, and must be installed as a systemd service.
There is a provided systemd service file.

This also depends on having a PostgreSQL database for storing versioning information.
A schema sql file is also provided.

## How do I use this

First, clone the repo and build the binary using ```go build -o cd_watcher ./cmd```, possibly
adding other parameters for your deploy environment.

Next, push the binary up to the server you are using, along-side the systemd service file.

Finally, perform all the configuration steps set forth in the following section.
Most of the configuration steps have sensible defaults, but the credentials section
cannot have a default.

To run a deploy, push a tarball up to the configured deploy directory, and then run
```cd_watcher deploy```.<br>
This scans the upload directory for any tarballs, decompresses them, puts them in
the releases directory, and updates the symlink for the currently deployed server.
Note: The tarballs are expected to be gzip compressed (created with tar -z)

If multiple tarballs are found, they are decompressed and added as releases in the
order that the go filesystem libraries find them. Only the last release is symlinked
and health checked.

It then runs two user supplied scripts, if they exist:
 - Restart (must exist)
 - Health Check

If either the health check script or the restart script returns an error code (exit
code not equal to 0), then a rollback will automatically occur

To run a rollback, just run ```cd_watcher rollback [number]```<br>
This rolls back a specific number of versions, if previous versions exist
and clears entries in the releases directory and database, updates the symlink
for the currently deployed server, and then restarts and health checks the server.

## Examples

Example configuration files can be found in the examples directory

## Configuration

### Unit File

It is recommended that you run this deploy unit as a separate user.<br>
A pair of systemd unit files (one deploy, one rollback) with all of these
parameters set can be found in the resources directory.

By default, it is set to the "cd_watcher" user, but that can be changed.

Similarly, the default working directory is the default home directory of
the "cd_watcher" user (/home/cd_watcher). Once again, this can be changed.

Finally, the path of the binary is set to be "/home/cd_watcher/bin/cd_watcher"
but this can be changed as well.

### Database

This runs using PostgreSQL as it's database for version management. I might eventually shift to
something like SQLite, but for now, it's PostgreSQL.<br>
An SQL file containing the up migration for the necessary tables can be found in the
resources directory.

### Credentials

All things in this section need to be [systemd credentials](https://www.freedesktop.org/software/systemd/man/latest/systemd.exec.html?__goaway_challenge=meta-refresh&__goaway_id=3c45b19eb29e9844aecce3a903c65bfd&__goaway_referer=https%3A%2F%2Fsystemd.io%2F#LoadCredential=ID:PATH).
 - pg_url: URL for the database connection. Always necessary.
 - email_parameters: JSON blob of data for configured email login type. Only necessary if
email configuration is not null

### JSON

You must put a ```config.json``` file in the working directory of the watcher.
All paths are relative to the working directory.

#### Releases Directory

```"release_dir"```<br>
Where are the releases going to go. Default: ```"releases"```
Note: Within the releases directory, there will be a symlink called ```current```.
This is the symlink that gets automatically updated.

#### Upload Directory

```"upload_dir"```<br>
Where are the tarballs going to be deployed to. Default: ```"uploads"```

#### Reload Script

```"reload"```<br>
The script to run to reload the service. Default ```"scripts/reload.sh"```

#### Health Check Script

```"health_check"```<br>
The script to run to run a health check on the service. Default ```""```<br>
Note: An empty string here means that the health check shouldn't be run

#### Emails

```"email_conf"```<br>
Object defining configuration for emails. Default ```null```<br>
Note: ```null ``` means that no emails will be sent for deploys. More information on
email configuration in the next section

## Email Configuration

Having some level of observability into what is happening is nice, so emails will get
sent out based on what is going on. It also handles threading automatically, so you won't
get 50 disparate emails, everything will just be in one email chain with replies baked in.

Under the hood, this uses [go-smtp](github.com/emersion/go-smtp), so you can check that out
for more information about some of the configuration options

### Email Events

There are 5 main events that you can listen to with emails:
 - Deploy started (```"deploy_start"```)
 - Deploy finished (```"deploy_finish"```)
 - Rollback started (```"rollback_start"```)
 - Rollback finished (```"rollback_finish"```)
 - Health check failed (```"health_check_fail"```)

### Email Host

```"host"```<br>
String defining the host and port for the SMTP server you are using.<br>
Note: No default value. Email client will force TLS for transmitting the
emails, so choose an SMTP port you can use TLS over

### Emailer
```"emailer"```<br>
String defining the email address of the sender of the emails.<br>
Note: No default value. Login info must be valid to send emails as the emailer.

### Login

```"login"```<br>
Login type used for the server.<br>
Note: No default value. Email client will authenticate the connection with
the SMTP server, so provide credentials. This is where the email_parameters
credential comes in.<br>
Under each heading will be a list of parameters that must be in the parameters
JSON blob<br>

#### Anonymous Login

```"trace"```: string

#### External Login

```"identity"```: string

#### OAuth Bearer Login

```"username"```: string<br>
```"token"```: string <br>
```"host"```: string<br>
```"port"```: integer

#### Plain Login
```"identity"```: string. If empty, assumed the same as ```"username"```<br>
```"username"```: string<br>
```"password"```: string

### Email Recipients
```"recipients"```: Array of objects, each with the following properties.<br>
 - ```"email"```: Email Address of the recipient
 - ```"events"```: Array of objects with the following properties
    - ```"name"```: The name of the event (see above)
    - ```"bcc"```: Whether you want the person to be BCC'd the email rather than sent it
