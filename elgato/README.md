# elgato

A command-line tool to toggle [Elgato Key Light](https://www.elgato.com/us/en/p/key-light) devices on your local network.

The tool uses mDNS (via zeroconf) to automatically discover the first Elgato Light on your network, then toggles it on or off.

## Usage

```
elgato            # Toggle the light on or off
elgato --help     # Show help message
elgato --version  # Show version info
```

## Building

```
make build
```

Cross-compile for other platforms:

```
make mac          # macOS (arm64 + amd64)
make linux        # Linux (arm64 + amd64)
make platforms    # All of the above
```

## Install

```
make install
```

This installs the binary to `/usr/local/bin`.

## How It Works

1. Discovers Elgato lights on the local network by browsing for `_elg._tcp` mDNS services.
2. Queries the light's HTTP API (port 9123) for its current state.
3. Sends a PUT request to flip the state (on → off, off → on).
