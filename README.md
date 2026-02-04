## ProxyChan

- ProxyChan is a lightweight, authenticated SOCKS5 proxy written in Go, designed to act as a controlled egress point for network traffic.

### Why ProxyChan exists
ProxyChan is built around one idea:

- Separate who can connect from where traffic exits.

It is for people who need controlled egress, not anonymity-by-default.

It sits between:
- raw SOCKS proxies (no control, no safety)
- full VPNs (heavy, opaque, all-or-nothing)

Typical use cases:
- exposing a SOCKS proxy safely to a LAN or remote users
- routing selected traffic through Tor without forcing Tor system-wide
- giving multiple users controlled outbound access from one machine
- lab, homelab, and security testing environments
- situations where VPNs are unnecessary or undesirable

ProxyChan is intentionally boring by design.
If traffic flows, it’s because a rule allows it.

This lets you:
 
- centralize outbound traffic
 
- control access with auth
 
- route traffic through Tor without running Tor everywhere
 
- use the proxy locally, on a LAN, or remotely
 
- avoid VPN complexity when SOCKS is enough

### Features

- SOCKS5 proxy (RFC-compliant)

- Username/password authentication

- Automatic auth enforcement

- Direct or Tor-based egress

- Optional dynamic proxy chaining

- Safe concurrent handling

- Clean shutdown via signals

- Source IP whitelist (client access control)

- Destination blacklist (egress control)

- Live policy reload (no restart required)

- SQLite-backed state shared between service and CLI

### System service installation:

- Linux (systemd)

- macOS (launchd)

- Windows (Service Manager)
```
# install
sudo ./proxychan --flag1 --flag2 --flag3 .etc install-service

# control
sudo systemctl start proxychan
sudo systemctl stop proxychan
sudo systemctl restart proxychan
sudo systemctl status proxychan

# remove
sudo ./proxychan remove-service
```

## Quick start
#### Build
```
go build -o proxychan
```
#### Run locally (no auth)
```
./proxychan -listen 127.0.0.1:1080
```
#### Add a user
```
./proxychan add-user
```

#### Run on all interfaces (auth required)
```
./proxychan -listen 0.0.0.0:1090 -mode tor
```

## Using the proxy

#### curl
```
curl --socks5 user:pass@HOST:PORT https://example.com
```

#### Firefox
```
Network Settings → Manual Proxy

SOCKS5

Enable Proxy DNS when using SOCKS v5

Enter credentials when prompted

Service installation (recommended)

ProxyChan can install itself as a native OS service.
```

#### Linux / macOS
```
sudo ./proxychan install-service \
  -listen 0.0.0.0:1090 \
  -mode tor
```

#### Windows (Admin shell)
```
proxychan.exe install-service -listen 0.0.0.0:1090 -mode tor
```

## After installation:

- The proxy runs in the background

- Logs go to the OS service logger

- Lifecycle is managed by the OS (start/stop/restart)

## Authentication model

- Binding to localhost → authentication not required
- Binding to non-local addresses → authentication enforced automatically

#### This prevents accidental open proxies while keeping local usage simple.

## Destination control (egress)

Outbound connections can be blocked by destination:

- IP address
- CIDR range
- Exact domain
- Domain suffix (e.g. .example.com)

#### Rules are applied before dialing out.
#### If a destination is blocked, no outbound connection is made.

## Access & policy model

- Source whitelist:
  Controls which client IPs are allowed to connect.

- Destination blacklist:
  Controls where traffic is allowed to go (IP, CIDR, domain, domain suffix).

#### These policies are enforced server-side and apply to all users equally.

## Visibility & privacy

### In Tor mode

- Proxy operator sees who connects
- Destinations are reached via Tor
- Destination metadata is hidden from the operator

### In direct mode:

- Proxy operator can observe destination metadata (IP/SNI)

- Payloads remain encrypted (HTTPS)

#### Choose the mode that fits your threat model.

## Management commands

### User management
- add-user
- del-user
- list-users
- list-user
- activate-user / deactivate-user
- activate-all / deactivate-all

### Source whitelist (client IPs)
- allow-ip
- block-ip
- del-ip
- list-whitelist
- clear-whitelist

### Destination blacklist (egress)
- block-dest
- allow-dest
- del-dest
- list-blacklist
- clear-blacklist

## What ProxyChan is not

- Not a VPN

- Not a packet sniffer

- Not a traffic analyzer

- Not a firewall

- Not a policy engine for inbound traffic


### It is a deliberate, minimal egress proxy.

## Project philosophy

- No daemon hacks

- No shell-script glue

- OS-native service management

- Explicit behavior over magic

- Policy is explicit and observable

# License

 MIT
