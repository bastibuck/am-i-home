# am-i-home-cli

CLI utility for checking which devices are (or were) connected to a Vodafone HomeStation router. It talks to the router's web interface, logs in and prints device data in a terminal-friendly table or checks for a certain device.

## Requirements
- Go 1.21+ (any modern Go toolchain should work)
- Network access to your router (default: http://192.168.0.1)

## Installation
```bash
go build -o am-i-home ./cmd/am-i-home
```
Or run directly with `go run ./cmd/am-i-home`.

## Credentials & Secrets
The CLI needs an admin username/password for your HomeStation. Password resolution happens in this order:
1. `-pass` flag
2. `AM_I_HOME_ROUTER_PASS` environment variable
3. `.env` file in the current working directory (simple `KEY=VALUE` pairs)
4. Interactive prompt (only if stdin is a TTY)

To avoid committing secrets, create a local `.env` file with `AM_I_HOME_ROUTER_PASS`. A template is available in `.env.example`.

## Usage
Flags must be provided before the command. The most common flags are:
- `-router` (default `http://192.168.0.1`)
- `-user` (default `admin`)
- `-pass` (see resolution order above)

Commands:
- `list` &mdash; print all currently active devices
- `list-all` &mdash; print every device the router has ever seen
- `check <MATCHER>` &mdash; return `true`/`false` depending on whether a matcher (MAC/hostname/IP) is active

Examples:
```bash
# show active devices using defaults
am-i-home list

# check if a device with hostname "work-laptop" is active, prompting for password
am-i-home -router http://192.168.1.1 check work-laptop
```

Exit codes:
- `0` matcher found
- `1` matcher not found
- `2` error (bad flag usage, login failure, etc.)
