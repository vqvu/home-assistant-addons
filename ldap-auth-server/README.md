# Home Assistant Add-on: LDAP Authentication Server

![Supports aarch64 Architecture][aarch64-shield] ![Supports amd64
Architecture][amd64-shield] ![Supports armhf Architecture][armhf-shield]
![Supports armv7 Architecture][armv7-shield] ![Supports i386
Architecture][i386-shield]

This add-on runs a simple HTTP server that can authenticate users against an
LDAP server. It is meant to be used in conjunction with Home Assistant's
[Command Line authentication
provider](https://www.home-assistant.io/docs/authentication/providers/#command-line),
so it only implements the bare minimum functionality to support that.

This add-on is as an alternative to the shell script at
<https://github.com/bob1de/ldap-auth-sh> for containerized Home Assistant
installs. Unlike `ldap-auth-sh`, all LDAP functionality is wrapped in the add-on
and exposed via an HTTP endpoint. That way, only `curl` is needed from within
the Home Assistant container.

[aarch64-shield]: https://img.shields.io/badge/aarch64-yes-green.svg
[amd64-shield]: https://img.shields.io/badge/amd64-yes-green.svg
[armhf-shield]: https://img.shields.io/badge/armhf-yes-green.svg
[armv7-shield]: https://img.shields.io/badge/armv7-yes-green.svg
[i386-shield]: https://img.shields.io/badge/i386-yes-green.svg

## Development

### LLDAP server

Starting the local LLDAP server

* Using the provided script:
  * The script generates a local root CA and a server certificate (if missing)
    under data/tls relative to the script location, then starts the server via
    Docker Compose from that directory. It does not depend on your current
    working directory.
  * Run:

    ```bash
    ./lldap_server/start_lldap.sh
    ```

* Using VS Code task:
  * Open the Command Palette → “Tasks: Run Task” → “Start LLDAP”.
  * This task invokes the same script so certificates are created under
    `lldap_server/data/tls` and the server is started with Docker Compose from
    that folder.

The LLDAP UI will be available at <http://localhost:17170> and is has an `admin`
user with password `adminpassword`.

Notes

* The generated certificates are for local development and include SANs for
  `localhost` and `127.0.0.1`.
* If you need to regenerate certificates, delete the files in
  `lldap_server/data/tls` and rerun the task or script.

### Starting a local Home Assistant

Follow instructions at <https://developers.home-assistant.io/docs/add-ons/testing/>.

### Install the add-on

Comment out the `image` key from the `config.yaml` to force a local build.

Use the following config:

```yaml
# Replace the IP address with the IP address of the LLDAP host.
ldap_server_url: ldaps://172.17.0.2:6360
bind_dn_template: cn=%s,ou=people,dc=example,dc=com
debug_mode: true
```

Enable port forwarding for port 8080. Call the server with

```bash
username=test password=testtest bash ./ldap_auth_command.sh
```
