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

Next, push the binary up to the server you are using, along-side the systemd service files.

Finally, perform all the configuration steps set forth in the following section.
Most of the configuration steps have sensible defaults, but the credentials section
cannot have a default.

To run a deploy, push a bundle up to the configured deploy directory, and then run
```cd_watcher deploy```.<br>
This scans the upload directory for any bundles, and runs the unpack script for
each bundle it finds.<br>
This unpack script gets run with the path to the bundle as the first argument.
It also sets the working directory to the created release directory, so you
can unpack the bundle here.<br>
After the unpack script runs, the symlink for the currently deployed server gets
updated.

If multiple bundles are found, the unpack script is run in the order the Go standard
library returns the bundles in, but only the last one updates the symlink

It then runs two user supplied scripts:
 - Restart
 - Health Check (optional)

If either the health check script or the restart script returns an error code (exit
code not equal to 0), then a rollback will automatically occur.

To run a rollback, just run ```cd_watcher rollback [number]```<br>
This rolls back a specific number of versions, if previous versions exist
and clears entries in the releases directory and database, updates the symlink
for the currently deployed server, and then restarts and health checks the server.

## Examples

Example configuration files can be found in the examples directory

## Scripts

There are a few user defined scripts for this project that makes it nice and
configurable, but there are a few restrictions for the scripts.

1. All scripts are run using ```/usr/bin/bash```
2. Keep script lines short. If output lines are excessively long, the entirety of
the command output could be thrown away
3. Script stdout and stderr is logged for revisiting if things break. Don't output
credentials or anything you wish to keep secret
4. The running waits for stdout and stderr to be closed. That means if you spawn a
background process but still pipe to stdout or stderr, the main process will wait
for those background processes to finish.

## Configuration

### Unit File

It is recommended that you run this deploy unit as a separate user.<br>
A pair of systemd unit files (one deploy, one rollback) with all of these
parameters set can be found in the resources directory, with the email
credential line commented out<br>

By default, it is set to the "cd_watcher" user, but that can be changed.

Similarly, the default working directory is the default home directory of
the "cd_watcher" user (/home/cd_watcher). Once again, this can be changed.

The unit files also have a file lock by default so if something is happening,
then another service starting will wait until the first service finishes

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
 - email_login: JSON blob of data for configured email login type. Only necessary if
email configuration is not null

These credentials will be read using two environment variables:
 - PG_URL_NAME: the name of the credential file that pg_url will be read from
 - EMAIL_LOGIN_NAME: the name of the credential file that email_login will be read from

You shouldn't put the credential directory in front, that will be appended in the
program itself.<br>
If the environment variables are empty (or unset), it will assume that the credentials
are just named 'pg_url' and 'email_login'

The environment variables are as necessary as their corresponding credentials.<br>
You might ask "Why put the credential name in an environment variable?"<br>
Great Question. Because the way systemd works is that when you create the credential,
it puts an ID with that credential. But, you might want to put a file extension with
the credential, which will then mess with the ID. So, you get to decide the ID and other
stuff and just tell the program which file to read.<br>

### JSON

You must put a ```config.json``` file in the working directory of the watcher.
All paths are relative to the working directory.

#### Releases Directory

```"release_dir"```<br>
Where are the releases going to go. Default: ```"releases"```<br>
Note: Within the releases directory, there will be a symlink called ```current```.
This is the symlink that gets automatically updated.

#### Upload Directory

```"upload_dir"```<br>
Where are the bundles going to be deployed to. Default: ```"uploads"```

#### Unpack Script

```"unpack"```<br>
The script to run to unpack the bundle. Default: ```"scripts/unpack.sh"```

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
Note: ```null``` means that no emails will be sent. More information on email
configuration in the next section

## Email Configuration

Having some level of observability into what is happening is nice, so emails will get
sent out based on what is going on. It also handles threading automatically, so you won't
get 50 disparate emails, everything will just be in one email chain with replies baked in.

Under the hood, this uses [go-smtp](github.com/emersion/go-smtp), so you can check that out
for more information about some of the configuration options

### Email Events

There are 4 main events that you can listen to with emails:
 - Deploy started (```"deploy_start"```)
 - Deploy finished (```"deploy_finish"```)
 - Rollback started (```"rollback_start"```)
 - Rollback finished (```"rollback_finish"```)

Finish emails will reply to their corresponding start email, new start
emails reply to the previous start email.<br>
If something went wrong (health check fail, other error) a finish email
will still be sent, it will just say there was an error<br>

There are 5 pseudo-events you can listen to, that are just combinations
of the 4 main events:
 - All (```"all"```)
 - Started (```"start"```)
 - Finished (```"finish"```)
 - Deploy (```"deploy"```)
 - Rollback (```"rollback"```)

It should be pretty obvious which events those all correlate to.<br>
These exist so you don't have to specify each command separately with
the same BCC rules

### Email Host

```"host"```<br>
String defining the host and port for the SMTP server you are using.<br>
Note: No default value. Email client will force TLS for transmitting the
emails, so choose an SMTP port you can use TLS over

### Emailer
```"emailer"```<br>
String defining the email address of the sender of the emails.<br>
Note: No default value. Login info must be valid to send emails as the emailer.

### Message ID Domain
```"message_id_domain"```<br>
String defining the domain of the message id for emails.<br>
Note: No default value. Domain must be valid for the emailer.
Message IDs are generated from the timestamp, and 36 bytes of
entropy. They are almost certainly guaranteed to be unique.<br>

### Login

```"login"```<br>
String value login type used for the server. "anonymous", "external", "oauth", or "plain"<br>
Note: No default value. Email client will authenticate the connection with
the SMTP server, so provide credentials. This is where the email_login
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
