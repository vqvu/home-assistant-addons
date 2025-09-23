# Home Assistant Add-on: LDAP Auth Server

## How to use

1. Install the add-on.
2. Configure `LDAP Server URL` and `Bind DN Template` to match your LDAP
    server.
3. (Optionally) Configure `Search Base DN` and `Search Filter Template` if you
    want to hide some users from Home Assistant. For more advanced configuration
    options, see the [Advanced configurations](#advanced-configurations)
    section.
4. Copy the
    [ldap_auth_command.sh](https://github.com/vqvu/home-assistant-addons/ldap-auth-server/ldap_auth_command.sh)
    file to your Home Assistant config directory.
5. Add a custom [Command Line authentication
    provider](https://www.home-assistant.io/docs/authentication/providers/#command-line)
    to your configuration to call that script.

    ```yaml
    homeassistant:
      auth_providers:
        - type: command_line
          command: /config/ldap_auth_command.sh
          args:
            # Provide the hostname of the add-on as the first argument. You can
            # the hostname on the add-on's Info page.
            - 7860403f-ldap-auth-server
          meta: true
        # Optionally add the homeassistant provider as a fallback if you're
        # concerned about a failed LDAP server locking you out of Home
        # Assistant.
        # - type: homeassistant
    ```

## Limitations

These limitations are current as of Home Assistant v2023.2.

1. The users created via the `command_line` provider are different from the
    ones created via the `homeassistant` provider, even if their usernames are
    the same. This means you will lose all user configurations when adopting a
    new auth provider.
2. Users created by `command_line` provider are all Administrators by default.
    If you don't want this, you can update the user configuration manually in
    the UI. Don't forget to restart your Home Assistant afterwards.

    While it is now possible to set the HASS group during user authentication,
    this repo does not currently implement it.
3. This add-on is only tested with an [LLDAP
    server](https://github.com/nitnelave/lldap), so it is possible (though
    probably unlikely) that it doesn't work with other types of LDAP servers for
    one reason or another.

## Advanced configurations

### Adding a custom CA root

If your LDAP server's certificate that is signed by a CA that is not trusted by
default, you can specify a CA chain in PEM format in a file under the directory
`/addon_configs/7860403f-ldap-auth-server`. Then specify the
`Server Root CAs File` config parameter.

Use a local path. If you put your file under
`/addon_configs/7860403f-ldap-auth-server/ca.pem`, the config should be
`ca.pem`.

## How it works

The add-on starts an HTTP server that accepts a username and password pair,
authenticates the user against an LDAP server, and returns user metadata in a
form that Home Assistant understands.

To authenticate a user, the add-on constructs the user's DN using the `Bind DN
Template`. Then it uses the password to bind to that DN. If successful, it
searches for the user record using the `Search Base DN` and `Search Filter
Template`. If the user is found, it returns the user's metadata. Otherwise, it
returns a 4xx status code.

This procedure is meant to simplify configuration. It avoids the need to
configure a special bind user for the add-on (as in common in LDAP
integrations). The downside is that it establishes a new connection each time.
