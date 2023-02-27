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
