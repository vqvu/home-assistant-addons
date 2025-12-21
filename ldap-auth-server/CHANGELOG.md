<!--
https://developers.home-assistant.io/docs/add-ons/presentation#keeping-a-changelog
-->

# Changelog

## [Unreleased]

## 0.3.0

### Added

* Improve documentations.
* Update base images to 3.23.
* Update `ldap_auth_command.sh` to hard-code safer defaults for new users.

### Fix

* Add `/config/**` to the apparmor profile.

## 0.2.0

### Added

* Add `Server Root CAs File` and `Disable Server Certificate Validation` parameters.

### Fixed

* Setting an empty `Search Base DN` and `Search Filter Template` should work now.

## 0.1.0

### Added

* Initial release
