# Local Setup Guide

This guide covers running Valhalla directly on your local machine without Docker or Kubernetes.

## Prerequisites

- **Data.nx file** - See [Installation Guide](Installation.md) for conversion instructions
- **MapleStory v28 client** - See [Installation Guide](Installation.md) for download
- **MySQL 5.7+** or **MariaDB** database server
- **Go 1.25+** (only if building from source)

## Quick Start

### Step 1: Download Valhalla

#### Option A: Pre-built Binaries (Recommended)

Download the latest release for your platform from the [releases page](https://github.com/Hucaru/Valhalla/releases).

Each release includes:
- Valhalla server binary
- Required JSON data files (drops.json, reactors.json, reactor_drops.json)
- Sample configuration files (config_*.toml)
- LICENSE and README

Extract the archive to your desired location.

#### Option B: Build from Source

See [Building.md](Building.md) for instructions on building from source.

### Step 2: Set Up Database

1. **Install MySQL/MariaDB** if not already installed:
   - Windows: [MySQL Installer](https://dev.mysql.com/downloads/installer/)
   - Linux: `sudo apt-get install mysql-server` or `sudo yum install mariadb-server`
   - macOS: `brew install mysql`

2. **Start the database service**:
   ```bash
   # Linux
   sudo systemctl start mysql
   
   # macOS
   brew services start mysql
   ```

3. **Create the database and import schema**:
   ```bash
   mysql -u root -p
   ```
   
   Then in the MySQL prompt:
   ```sql
   CREATE DATABASE maplestory;
   exit;
   ```
   
   Import the schema:
   ```bash
   mysql -u root -p maplestory < maplestory.sql
   ```

### Step 3: Prepare Data.nx

1. Convert your Data.wz file to Data.nx format (see [Installation Guide](Installation.md))
2. Place the `Data.nx` file in the Valhalla directory

Your directory should look like:
```
Valhalla/
├── Valhalla (or Valhalla.exe)
├── Data.nx
├── drops.json
├── reactors.json
├── reactor_drops.json
├── config_login.toml
├── config_world.toml
├── config_channel_1.toml
├── config_channel_2.toml
├── config_channel_3.toml
├── config_cashshop.toml
└── scripts/
```

### Step 4: Configure Server

Edit the configuration files to match your setup. For local development, the defaults should work:

#### config_login.toml
```toml
[database]
address = "127.0.0.1"
port = "3306"
user = "root"
password = "your_password"
database = "maplestory"

[login]
clientListenAddress = "0.0.0.0"
clientListenPort = "8484"
serverListenAddress = "0.0.0.0"
serverListenPort = "8485"
withPin = false
autoRegister = true  # Set to true for easy testing
packetQueueSize = 512
latency = 0
jitter = 0
```

#### config_world.toml
```toml
[database]
address = "127.0.0.1"
port = "3306"
user = "root"
password = "your_password"
database = "maplestory"

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

#### config_channel_1.toml (and _2, _3, etc.)
```toml
[database]
address = "127.0.0.1"
port = "3306"
user = "root"
password = "your_password"
database = "maplestory"

[channel]
worldAddress = "127.0.0.1"
worldPort = "8584"
listenAddress = "0.0.0.0"
listenPort = "8685"  # 8686 for channel 2, 8687 for channel 3, etc.
clientConnectionAddress = "127.0.0.1"
packetQueueSize = 512
maxPop = 250
latency = 0
jitter = 0
```

**Note**: Update the password in all config files to match your MySQL root password.

### Step 5: Start the Servers

Start each server component in order. Use separate terminal windows/tabs for each:

**Terminal 1 - Login Server:**
```bash
# Windows
Valhalla.exe -type login -config config_login.toml

# Linux/macOS
./Valhalla -type login -config config_login.toml
```

**Terminal 2 - World Server:**
```bash
# Windows
Valhalla.exe -type world -config config_world.toml

# Linux/macOS
./Valhalla -type world -config config_world.toml
```

**Terminal 3 - Channel Server 1:**
```bash
# Windows
Valhalla.exe -type channel -config config_channel_1.toml

# Linux/macOS
./Valhalla -type channel -config config_channel_1.toml
```

**Terminal 4 (Optional) - Additional Channels:**
```bash
# For channel 2
./Valhalla -type channel -config config_channel_2.toml

# For channel 3
./Valhalla -type channel -config config_channel_3.toml
```

**Terminal 5 (Optional) - Cash Shop Server:**
```bash
# Windows
Valhalla.exe -type cashshop -config config_cashshop.toml

# Linux/macOS
./Valhalla -type cashshop -config config_cashshop.toml
```

### Step 6: Connect with Client

1. Launch your MapleStory v28 client (see [Installation Guide](Installation.md) for client download)
2. The client should connect to `127.0.0.1:8484`
3. With `autoRegister = true`, you can login with any username/password to create a new account

## Managing the Server

### Starting/Stopping

To stop a server, press `Ctrl+C` in its terminal window.

Start servers in this order:
1. Login Server
2. World Server
3. Channel Server(s)
4. Cash Shop Server (optional)

Stop servers in reverse order for clean shutdown.

### Logs

Server logs are printed to stdout/stderr. To save logs to a file:

```bash
# Linux/macOS
./Valhalla -type login -config config_login.toml > login.log 2>&1

# Windows PowerShell
.\Valhalla.exe -type login -config config_login.toml > login.log 2>&1
```

### Using a Process Manager

For easier management, use a process manager:

#### Linux - systemd

Create a service file `/etc/systemd/system/valhalla-login.service`:

```ini
[Unit]
Description=Valhalla Login Server
After=mysql.service

[Service]
Type=simple
User=valhalla
WorkingDirectory=/path/to/valhalla
ExecStart=/path/to/valhalla/Valhalla -type login -config config_login.toml
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Then:
```bash
sudo systemctl daemon-reload
sudo systemctl enable valhalla-login
sudo systemctl start valhalla-login
```

Repeat for world, channel, and cashshop servers.

#### Windows - NSSM

Use [NSSM (Non-Sucking Service Manager)](https://nssm.cc/):

```cmd
nssm install ValhallaLogin "C:\path\to\Valhalla.exe" "-type login -config config_login.toml"
nssm start ValhallaLogin
```

#### macOS - launchd

Create a plist file `~/Library/LaunchAgents/com.valhalla.login.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.valhalla.login</string>
    <key>ProgramArguments</key>
    <array>
        <string>/path/to/Valhalla</string>
        <string>-type</string>
        <string>login</string>
        <string>-config</string>
        <string>config_login.toml</string>
    </array>
    <key>WorkingDirectory</key>
    <string>/path/to/valhalla</string>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>
```

Then:
```bash
launchctl load ~/Library/LaunchAgents/com.valhalla.login.plist
```

## Monitoring

### Metrics

Valhalla exposes Prometheus metrics on port 9000 by default (configurable with `-metrics-port`).

View metrics at: `http://localhost:9000/metrics`

Metrics include:
- Channel population
- Server thread count
- Memory usage
- Monster kill rate
- Active trades, minigames, NPC interactions
- Number of parties

### Setting Up Prometheus + Grafana

1. **Install Prometheus**:
   - Download from [prometheus.io](https://prometheus.io/download/)
   - Configure `prometheus.yml`:
   ```yaml
   scrape_configs:
     - job_name: 'valhalla'
       static_configs:
         - targets: ['localhost:9000']
   ```

2. **Install Grafana**:
   - Download from [grafana.com](https://grafana.com/grafana/download)
   - Add Prometheus as a data source
   - Create dashboards for your metrics

## Troubleshooting

### Can't Connect to Database

**Error**: `Error connecting to database`

**Solutions**:
- Verify MySQL is running: `systemctl status mysql` (Linux) or check Services (Windows)
- Check credentials in config files
- Test connection: `mysql -u root -p -h 127.0.0.1`

### Client Can't Connect

**Error**: Client shows "Unable to connect to server"

**Solutions**:
- Ensure login server is running on port 8484
- Check firewall settings: `sudo ufw allow 8484/tcp` (Linux)
- Verify client is patched for localhost (see [Installation Guide](Installation.md))

### Missing Data.nx Error

**Error**: `Failed to load Data.nx`

**Solutions**:
- Ensure Data.nx is in the same directory as the executable
- Verify the file was converted correctly and is not corrupted
- Check file permissions

### Port Already in Use

**Error**: `bind: address already in use`

**Solutions**:
- Check if another instance is running
- Find process using port: `netstat -ano | findstr :8484` (Windows) or `lsof -i :8484` (Linux/macOS)
- Change port in configuration file

### Server Crashes on Startup

**Solutions**:
- Check that all JSON files (drops.json, reactors.json, reactor_drops.json) are present
- Verify scripts/ directory exists
- Check terminal output for specific error messages
- Ensure Go version is 1.25+ if building from source

## Next Steps

- Configure server settings: [Configuration.md](Configuration.md)
- Build from source for development: [Building.md](Building.md)
- Deploy with Docker for easier management: [Docker.md](Docker.md)

## Performance Tips

For better performance on local setups:

1. **Use SSD storage** for database and Data.nx
2. **Allocate enough RAM** - Minimum 4GB, recommended 8GB+
3. **Optimize MySQL**:
   ```ini
   # /etc/mysql/my.cnf
   [mysqld]
   innodb_buffer_pool_size = 1G
   max_connections = 200
   ```
4. **Disable latency simulation** in config files (set latency=0, jitter=0)
