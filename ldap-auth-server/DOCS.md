# Home Assistant Add-on: LDAP Auth Server

## How to use

1. Install the add-on.
2. Configure `LDAP Server URL` and `Bind DN Template` to match your LDAP
    server.
3. (Optionally) Configure `Search Base DN` and `Search Filter Template` if you
    want to hide some users from Home Assistant.
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
