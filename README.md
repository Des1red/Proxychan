## ProxyChan

- ProxyChan is a lightweight, authenticated SOCKS5 proxy written in Go, designed to act as a controlled egress point for network traffic.

- It can run in:

- direct mode (traditional proxy)
 
- Tor mode (proxy → Tor → Internet)
 
- chained mode (dynamic SOCKS hop chaining)
 
- It supports per-user authentication, LAN / remote usage, and OS-native service installation.

### Why ProxyChan exists

#### ProxyChan is built around one idea:

- Separate who can connect from where traffic exits.
 
#### This lets you:
 
- centralize outbound traffic
 
- control access with auth
 
- route traffic through Tor without running Tor everywhere
 
- use the proxy locally, on a LAN, or remotely
 
- avoid VPN complexity when SOCKS is enough
 
- Think of it as SSH for network egress.

### Features

- SOCKS5 proxy (RFC-compliant)

- Username/password authentication

- Automatic auth enforcement when not bound to localhost

- Direct or Tor-based egress

- Optional dynamic proxy chaining

- Safe concurrent handling

- Clean shutdown via signals

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

- Authentication model

- localhost bind → no auth required

- non-local bind → auth enforced automatically

- This prevents accidental open proxies while keeping local development simple.

## Visibility & privacy

### In Tor mode:

- Proxy operator sees who connects, not where they go

- Destinations see Tor exit IPs

### In direct mode:

- Proxy operator can observe destination metadata (IP/SNI)

- Payloads remain encrypted (HTTPS)

#### Choose the mode that fits your threat model.

## What ProxyChan is not

- Not a VPN

- Not a packet sniffer

- Not a traffic analyzer

- Not a firewall

### It is a deliberate, minimal egress proxy.

## Project philosophy

- No daemon hacks

- No shell-script glue

- OS-native service management

- Explicit behavior over magic

If something runs, it’s because the OS runs it.

# License

 MIT