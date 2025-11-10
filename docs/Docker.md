# Docker Setup Guide

This guide covers running Valhalla using Docker and Docker Compose.

## Why Docker?

Docker provides several advantages:
- **Easy setup** - No need to install Go, MySQL, or manage dependencies
- **Consistent environment** - Same setup works across Windows, Linux, and macOS
- **Isolation** - Services run in containers without affecting your system
- **Easy scaling** - Add more channels by editing docker-compose.yml

## Prerequisites

- **Docker** - [Install Docker Desktop](https://www.docker.com/products/docker-desktop)
- **Docker Compose** - Included with Docker Desktop
- **Data.nx file** - See [Installation Guide](Installation.md) for conversion instructions

## Quick Start

### Step 1: Clone or Download Valhalla

```bash
git clone https://github.com/Hucaru/Valhalla.git
cd Valhalla
```

Or download the source code from the [releases page](https://github.com/Hucaru/Valhalla/releases).

### Step 2: Prepare Data.nx

1. Convert your Data.wz file to Data.nx format (see [Installation Guide](Installation.md))
2. Place the `Data.nx` file in the root Valhalla directory

### Step 3: Start the Services

```bash
docker-compose up -d
```

This will:
- Build the Valhalla Docker image
- Start MySQL database (port 3306)
- Start Login server (port 8484)
- Start World server (port 8584)
- Start Cash Shop server (port 8600)
- Start 2 Channel servers (ports 8685, 8686)
- Start Adminer web interface (port 8080) - for database management
- Start Prometheus (port 9090) - for metrics
- Start Grafana (port 3000) - for dashboards

### Step 4: Wait for Services to Initialize

Check the logs to ensure all services started successfully:

```bash
docker-compose logs -f
```

Look for messages indicating successful startup. Press `Ctrl+C` to exit logs.

### Step 5: Connect with Client

Launch your MapleStory v28 client (see [Installation Guide](Installation.md)). The client should connect to `127.0.0.1:8484`.

## Docker Compose Services

The default `docker-compose.yml` includes:

| Service | Port | Description |
|---------|------|-------------|
| `login_server` | 8484 | Handles authentication |
| `world_server` | 8584 | Manages world state |
| `cashshop_server` | 8600 | Cash shop functionality |
| `channel_server_1` | 8685 | Game channel 1 |
| `channel_server_2` | 8686 | Game channel 2 |
| `db` | 3306 | MySQL database |
| `adminer` | 8080 | Web-based database admin |
| `prometheus` | 9090 | Metrics collection |
| `grafana` | 3000 | Metrics visualization |

## Managing Services

### View Running Containers

```bash
docker-compose ps
```

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f login_server

# Last 100 lines
docker-compose logs --tail=100 channel_server_1
```

### Stop Services

```bash
docker-compose stop
```

### Start Services

```bash
docker-compose start
```

### Restart Services

```bash
# Restart all
docker-compose restart

# Restart specific service
docker-compose restart channel_server_1
```

### Stop and Remove Everything

```bash
docker-compose down
```

### Stop and Remove Everything Including Database Volume

**Warning**: This will delete all your database data!

```bash
docker-compose down -v
```

## Configuration

### Environment Variables

All configuration is done through environment variables in `docker-compose.yml`.

The file defines common environment variables using YAML anchors:

```yaml
x-common-env: &common_env
    VALHALLA_DATABASE_ADDRESS: "db"
    VALHALLA_DATABASE_PORT: "3306"
    # ... more variables
```

These are then applied to each service:

```yaml
services:
    login_server:
        environment:
            <<: *common_env
```

### Changing Configuration

Edit `docker-compose.yml` and modify the environment variables:

```yaml
x-common-env: &common_env
    # Change database password
    VALHALLA_DATABASE_PASSWORD: "mySecurePassword"
    
    # Change rates
    VALHALLA_WORLD_EXPRATE: "2.0"    # 2x EXP
    VALHALLA_WORLD_DROPRATE: "1.5"   # 1.5x Drop
    VALHALLA_WORLD_MESOSRATE: "1.5"  # 1.5x Mesos
```

After making changes:

```bash
docker-compose down
docker-compose up -d
```

See [Configuration.md](Configuration.md) for all available options.

## Adding More Channels

To add additional channels:

1. Edit `docker-compose.yml` and add a new service:

```yaml
    channel_server_3:
        build:
            context: .
            dockerfile: Dockerfile
        container_name: channel-server-3
        command: ["/app/Valhalla", "-type", "channel"]
        restart: unless-stopped
        volumes:
            - ./docker/docker_config_channel_3.toml:/app/docker/docker_config_channel.toml
            - ./Data.nx:/app/Data.nx
        ports:
            - 8687:8687
        depends_on:
            - world_server
        environment:
            <<: *common_env
            VALHALLA_CHANNEL_LISTENPORT: "8687"
```

2. Restart services:

```bash
docker-compose up -d
```

**Note**: Port numbers decrease by 1 for each additional channel (8685, 8686, 8687, etc.).

## External Access

To allow external connections (LAN or Internet):

### For Local Network (LAN)

1. Find your server's local IP:
   ```bash
   # Linux/macOS
   ip addr show
   
   # Windows
   ipconfig
   ```

2. Update `docker-compose.yml`:
   ```yaml
   VALHALLA_CHANNEL_CLIENTCONNECTIONADDRESS: "192.168.1.100"  # Your local IP
   VALHALLA_CASHSHOP_CLIENTCONNECTIONADDRESS: "192.168.1.100"
   ```

3. Restart services:
   ```bash
   docker-compose down
   docker-compose up -d
   ```

### For Internet Access

1. Set up port forwarding on your router:
   - Forward ports 8484, 8600, 8685, 8686, etc. to your server's local IP

2. Update `docker-compose.yml` with your public IP:
   ```yaml
   VALHALLA_CHANNEL_CLIENTCONNECTIONADDRESS: "203.0.113.1"  # Your public IP
   VALHALLA_CASHSHOP_CLIENTCONNECTIONADDRESS: "203.0.113.1"
   ```

3. Restart services

## Database Management

### Using Adminer (Web Interface)

Access Adminer at `http://localhost:8080`:
- **System**: MySQL
- **Server**: db
- **Username**: root
- **Password**: password (or your custom password)
- **Database**: maplestory

### Using MySQL Command Line

```bash
# Connect to database
docker-compose exec db mysql -u root -ppassword maplestory

# Run SQL file
docker-compose exec -T db mysql -u root -ppassword maplestory < backup.sql

# Backup database
docker-compose exec db mysqldump -u root -ppassword maplestory > backup.sql
```

## Monitoring

### Prometheus

Access Prometheus at `http://localhost:9090`

Query examples:
- `valhalla_channel_population` - Channel player count
- `valhalla_monster_kill_rate` - Monsters killed per second
- `go_memstats_alloc_bytes` - Memory usage

### Grafana

Access Grafana at `http://localhost:3000`

Default credentials:
- **Username**: admin
- **Password**: admin

1. Add Prometheus data source:
   - URL: `http://prometheus:9090`
   
2. Import or create dashboards for Valhalla metrics

See the main README for example dashboard screenshots.

## Volumes

Docker Compose creates persistent volumes:

```bash
# List volumes
docker volume ls | grep valhalla

# Inspect database volume
docker volume inspect valhalla_db-data

# Backup database volume
docker run --rm -v valhalla_db-data:/data -v $(pwd):/backup alpine tar czf /backup/db-backup.tar.gz -C /data .

# Restore database volume
docker run --rm -v valhalla_db-data:/data -v $(pwd):/backup alpine tar xzf /backup/db-backup.tar.gz -C /data
```

## Building Custom Images

### Build Image

```bash
docker build -t valhalla:custom .
```

### Use Custom Image in Docker Compose

Edit `docker-compose.yml`:

```yaml
services:
    login_server:
        image: valhalla:custom
        # Remove the 'build' section
```

## Troubleshooting

### Container Exits Immediately

**Check logs**:
```bash
docker-compose logs login_server
```

**Common causes**:
- Missing Data.nx file
- Database not ready yet
- Port already in use

### Can't Connect to Database

**Check database status**:
```bash
docker-compose ps db
docker-compose logs db
```

**Solutions**:
- Wait longer for database to initialize (30-60 seconds on first start)
- Check database credentials in environment variables

### Port Already in Use

**Error**: `Bind for 0.0.0.0:8484 failed: port is already allocated`

**Solutions**:
- Check for other running instances: `docker-compose ps`
- Check for processes using the port: `netstat -ano | findstr :8484` (Windows) or `lsof -i :8484` (Linux/macOS)
- Change port mapping in docker-compose.yml: `"8484:8484"` → `"8485:8484"`

### Client Can't Connect

**Solutions**:
1. Verify all services are running: `docker-compose ps`
2. Check login server logs: `docker-compose logs login_server`
3. Ensure client is patched for localhost (see [Installation Guide](Installation.md))
4. Check firewall settings

### High Memory Usage

**Limit memory per service**:

```yaml
services:
    channel_server_1:
        deploy:
            resources:
                limits:
                    memory: 512M
```

### Rebuilding After Code Changes

```bash
# Rebuild images
docker-compose build

# Recreate containers with new image
docker-compose up -d --force-recreate
```

## Docker Compose Commands Reference

```bash
# Start in foreground (see logs)
docker-compose up

# Start in background
docker-compose up -d

# Stop services
docker-compose stop

# Start stopped services
docker-compose start

# Restart services
docker-compose restart

# View logs
docker-compose logs -f

# Remove stopped containers
docker-compose down

# Remove containers and volumes
docker-compose down -v

# Rebuild images
docker-compose build

# Pull latest images
docker-compose pull

# Execute command in running container
docker-compose exec login_server sh

# View resource usage
docker stats
```

## Next Steps

- Configure server settings: [Configuration.md](Configuration.md)
- Set up Kubernetes deployment: [Kubernetes.md](Kubernetes.md)
- Learn about building from source: [Building.md](Building.md)

## Best Practices

1. **Use environment variables** for secrets and configuration
2. **Backup database regularly** using volume backups or mysqldump
3. **Monitor resource usage** with `docker stats`
4. **Keep images updated** by rebuilding periodically
5. **Use Docker volumes** for persistent data, not bind mounts
6. **Set resource limits** to prevent containers from consuming all system resources
7. **Use Docker secrets** for production deployments (instead of environment variables)
