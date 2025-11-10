# Configuration Guide

Valhalla can be configured using TOML configuration files or environment variables. This guide covers all available configuration options.

## Configuration Methods

### TOML Configuration Files

Each server type has its own configuration file:
- `config_login.toml` - Login server
- `config_world.toml` - World server
- `config_channel_<N>.toml` - Channel servers (one per channel)
- `config_cashshop.toml` - Cash shop server

### Environment Variables

For containerized deployments, configuration can be set via environment variables. Variable names follow this pattern:

```
VALHALLA_<SECTION>_<KEY>=value
```

For example:
```toml
[login]
clientListenAddress = "0.0.0.0"
```

Becomes:
```bash
VALHALLA_LOGIN_CLIENTLISTENADDRESS=0.0.0.0
```

## Command Line Flags

All server executables accept the following flags:

| Flag | Required | Description | Example |
|------|----------|-------------|---------|
| `-type` | Yes | Server type to start | `-type login`, `-type world`, `-type channel`, `-type cashshop` |
| `-config` | No | Path to TOML config file | `-config config_login.toml` |
| `-metrics-port` | No | Port for Prometheus metrics | `-metrics-port 9000` (default) |

### Example Commands

```bash
# Login server with custom config
./Valhalla -type login -config config_login.toml

# World server with default config and custom metrics port
./Valhalla -type world -metrics-port 9001

# Channel server
./Valhalla -type channel -config config_channel_1.toml
```

## Database Configuration

All server types require database configuration.

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `address` | string | Database server address | `127.0.0.1` |
| `port` | string | Database server port | `3306` |
| `user` | string | Database username | `root` |
| `password` | string | Database password | `password` |
| `database` | string | Database name | `maplestory` |

**Environment Variable Prefix**: `VALHALLA_DATABASE_`

### Example

```toml
[database]
address = "127.0.0.1"
port = "3306"
user = "root"
password = "password"
database = "maplestory"
```

## Login Server Configuration

Configuration section: `[login]`

| Parameter | Type | Description | Default | Env Variable |
|-----------|------|-------------|---------|--------------|
| `clientListenAddress` | string | Address to listen for client connections | `0.0.0.0` | `VALHALLA_LOGIN_CLIENTLISTENADDRESS` |
| `clientListenPort` | string | Port for client connections | `8484` | `VALHALLA_LOGIN_CLIENTLISTENPORT` |
| `serverListenAddress` | string | Address to listen for server connections | `0.0.0.0` | `VALHALLA_LOGIN_SERVERLISTENADDRESS` |
| `serverListenPort` | string | Port for server connections | `8485` | `VALHALLA_LOGIN_SERVERLISTENPORT` |
| `withPin` | bool | Enable PIN code requirement | `false` | `VALHALLA_LOGIN_WITHPIN` |
| `autoRegister` | bool | Auto-create accounts on login attempt | `false` | `VALHALLA_LOGIN_AUTOREGISTER` |
| `packetQueueSize` | int | Size of packet processing queue | `512` | `VALHALLA_LOGIN_PACKETQUEUESIZE` |
| `latency` | int | Simulated latency in milliseconds (for testing) | `0` | `VALHALLA_LOGIN_LATENCY` |
| `jitter` | int | Simulated jitter in milliseconds (for testing) | `0` | `VALHALLA_LOGIN_JITTER` |

### Auto-Register Feature

When `autoRegister = true`:
- New accounts are automatically created when users attempt to login with non-existent credentials
- Default values: gender=0, dob=1111111, eula=1, adminLevel=0, PIN="1111"
- **Security Note**: Only enable this for development or private servers. Disable for production.

### Example

```toml
[login]
clientListenAddress = "0.0.0.0"
clientListenPort = "8484"
serverListenAddress = "0.0.0.0"
serverListenPort = "8485"
withPin = false
autoRegister = false
packetQueueSize = 512
latency = 0
jitter = 0
```

## World Server Configuration

Configuration section: `[world]`

| Parameter | Type | Description | Default | Env Variable |
|-----------|------|-------------|---------|--------------|
| `message` | string | World server message displayed to players | `message` | `VALHALLA_WORLD_MESSAGE` |
| `ribbon` | int | Ribbon type (0=none, 1=event, 2=new, 3=hot) | `2` | `VALHALLA_WORLD_RIBBON` |
| `expRate` | float | Experience rate multiplier | `1.0` | `VALHALLA_WORLD_EXPRATE` |
| `dropRate` | float | Drop rate multiplier | `1.0` | `VALHALLA_WORLD_DROPRATE` |
| `mesosRate` | float | Mesos rate multiplier | `1.0` | `VALHALLA_WORLD_MESOSRATE` |
| `loginAddress` | string | Login server address | `127.0.0.1` | `VALHALLA_WORLD_LOGINADDRESS` |
| `loginPort` | string | Login server port | `8485` | `VALHALLA_WORLD_LOGINPORT` |
| `listenAddress` | string | Address to listen for connections | `0.0.0.0` | `VALHALLA_WORLD_LISTENADDRESS` |
| `listenPort` | string | Port to listen on | `8584` | `VALHALLA_WORLD_LISTENPORT` |
| `packetQueueSize` | int | Size of packet processing queue | `512` | `VALHALLA_WORLD_PACKETQUEUESIZE` |

### Ribbon Types

- `0` - No ribbon
- `1` - Event ribbon (yellow)
- `2` - New ribbon (green) 
- `3` - Hot ribbon (red)

### Example

```toml
[world]
message = "Welcome to Valhalla!"
ribbon = 2
expRate = 1.0
dropRate = 1.0
mesosRate = 1.0
loginAddress = "127.0.0.1"
loginPort = "8485"
listenAddress = "0.0.0.0"
listenPort = "8584"
packetQueueSize = 512
```

## Channel Server Configuration

Configuration section: `[channel]`

| Parameter | Type | Description | Default | Env Variable |
|-----------|------|-------------|---------|--------------|
| `worldAddress` | string | World server address | `127.0.0.1` | `VALHALLA_CHANNEL_WORLDADDRESS` |
| `worldPort` | string | World server port | `8584` | `VALHALLA_CHANNEL_WORLDPORT` |
| `listenAddress` | string | Address to listen for server connections | `0.0.0.0` | `VALHALLA_CHANNEL_LISTENADDRESS` |
| `listenPort` | string | Port for server connections | `8685` | `VALHALLA_CHANNEL_LISTENPORT` |
| `clientConnectionAddress` | string | Address clients connect to (external IP) | `127.0.0.1` | `VALHALLA_CHANNEL_CLIENTCONNECTIONADDRESS` |
| `packetQueueSize` | int | Size of packet processing queue | `512` | `VALHALLA_CHANNEL_PACKETQUEUESIZE` |
| `maxPop` | int | Maximum channel population | `250` | `VALHALLA_CHANNEL_MAXPOP` |
| `latency` | int | Simulated latency in milliseconds (for testing) | `0` | `VALHALLA_CHANNEL_LATENCY` |
| `jitter` | int | Simulated jitter in milliseconds (for testing) | `0` | `VALHALLA_CHANNEL_JITTER` |

### Important: Multiple Channels

Each channel requires its own configuration file and process. Channels are numbered starting from 1:
- Channel 1: `listenPort = 8685`
- Channel 2: `listenPort = 8686`
- Channel 3: `listenPort = 8687`
- And so on...

### Example

```toml
[channel]
worldAddress = "127.0.0.1"
worldPort = "8584"
listenAddress = "0.0.0.0"
listenPort = "8685"
clientConnectionAddress = "127.0.0.1"
packetQueueSize = 512
maxPop = 250
latency = 0
jitter = 0
```

## Cash Shop Server Configuration

Configuration section: `[cashshop]`

| Parameter | Type | Description | Default | Env Variable |
|-----------|------|-------------|---------|--------------|
| `worldAddress` | string | World server address | `127.0.0.1` | `VALHALLA_CASHSHOP_WORLDADDRESS` |
| `worldPort` | string | World server port | `8584` | `VALHALLA_CASHSHOP_WORLDPORT` |
| `listenAddress` | string | Address to listen for server connections | `0.0.0.0` | `VALHALLA_CASHSHOP_LISTENADDRESS` |
| `listenPort` | string | Port for server connections | `8600` | `VALHALLA_CASHSHOP_LISTENPORT` |
| `clientConnectionAddress` | string | Address clients connect to (external IP) | `127.0.0.1` | `VALHALLA_CASHSHOP_CLIENTCONNECTIONADDRESS` |
| `packetQueueSize` | int | Size of packet processing queue | `512` | `VALHALLA_CASHSHOP_PACKETQUEUESIZE` |
| `latency` | int | Simulated latency in milliseconds (for testing) | `0` | `VALHALLA_CASHSHOP_LATENCY` |
| `jitter` | int | Simulated jitter in milliseconds (for testing) | `0` | `VALHALLA_CASHSHOP_JITTER` |

### Example

```toml
[cashshop]
worldAddress = "127.0.0.1"
worldPort = "8584"
listenAddress = "0.0.0.0"
listenPort = "8600"
clientConnectionAddress = "127.0.0.1"
packetQueueSize = 512
latency = 0
jitter = 0
```

## Network Configuration Tips

### Local Development

For local testing:
```toml
clientConnectionAddress = "127.0.0.1"
```

### LAN/Remote Access

For LAN or internet access, set to your server's external IP:
```toml
clientConnectionAddress = "192.168.1.100"  # LAN
# or
clientConnectionAddress = "203.0.113.1"    # Public IP
```

### Docker/Kubernetes

See [Docker.md](Docker.md) and [Kubernetes.md](Kubernetes.md) for container-specific networking.

## Performance Tuning

### Packet Queue Size

Higher values can handle more simultaneous connections but use more memory:
- Small servers (1-50 players): `512`
- Medium servers (50-200 players): `1024`
- Large servers (200+ players): `2048`

### Latency Simulation

For development environments mimicking real-world conditions:
```toml
latency = 50    # 50ms base latency
jitter = 10     # ±10ms variation
```

**Note**: Always set to `0` for production servers.

## See Also

- [Installation Guide](Installation.md) - Setting up Valhalla
- [Local Setup](Local.md) - Running locally
- [Docker Setup](Docker.md) - Running with Docker
- [Kubernetes Setup](Kubernetes.md) - Kubernetes deployment
