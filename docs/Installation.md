# Installation Guide

This guide will walk you through setting up Valhalla, a MapleStory v28 private server emulator.

## Prerequisites

Before you begin, you'll need:
- **Data.nx file** - Required for all server components, especially channels
- A MapleStory v28 client
- A MySQL/MariaDB database

## Quick Navigation

Choose your installation method:
- [Local Setup](Local.md) - Run directly on your machine
- [Docker Setup](Docker.md) - Run using Docker Compose
- [Kubernetes Setup](Kubernetes.md) - Deploy to a Kubernetes cluster

For development work, see [Building from Source](Building.md).

For configuration options, see [Configuration Guide](Configuration.md).

## Converting Data.wz to Data.nx

All deployment methods require a `Data.nx` file. This file is generated from MapleStory's `Data.wz` file.

### Step 1: Locate Data.wz

After installing MapleStory v28:
1. Navigate to your MapleStory installation directory
2. Find the `Data.wz` file (typically in the root installation folder)

### Step 2: Convert to NX Format

Use the [go-wztonx-converter](https://github.com/ErwinsExpertise/go-wztonx-converter) tool to convert the WZ file to NX format.

#### Download the Converter

Download the pre-built binary for your platform from the [releases page](https://github.com/ErwinsExpertise/go-wztonx-converter/releases).

Available for:
- Windows (amd64)
- Linux (amd64)
- macOS (amd64/arm64)

#### Run the Conversion

```bash
# Windows
go-wztonx-converter.exe --server Data.wz

# Linux/macOS
./go-wztonx-converter --server Data.wz
```

This will generate a `Data.nx` file in the same directory.

### Step 3: Place Data.nx File

Copy the generated `Data.nx` file to your Valhalla installation directory:

- **Local setup**: Place in the root of your Valhalla directory
- **Docker setup**: Place in the root directory (it will be mounted into containers)
- **Kubernetes setup**: You'll need to create a ConfigMap or PersistentVolume (see [Kubernetes.md](Kubernetes.md))

## Setting Up the Client

### Download Client

A pre-patched MapleStory v28 client is available for localhost connections:

[Download v28 Client (no AES, patched, windowed mode, no IE check)](https://github.com/user-attachments/files/19866472/v28_noaes_patched_res_noie_mp.zip)

This client includes:
- No encryption (noaes)
- Pre-patched for localhost
- Window mode support
- IE check removed

### Alternative: Custom Client Hook

For more control, you can use a DLL hook to modify the client:

[MapleStory Client Hook](https://github.com/Hucaru/maplestory-client-hook)

This allows you to:
- Force localhost connections
- Enable windowed mode
- Bypass various client checks

## Database Setup

All installation methods require a MySQL database.

### Create Database

The server expects a database named `maplestory` (configurable).

1. Download the SQL schema:
   ```bash
   # In your Valhalla directory
   mysql -u root -p < maplestory.sql
   ```

2. Or use the docker-compose setup which automatically initializes the database.

## Next Steps

Choose your installation method:

- **[Local Setup](Local.md)** - Best for quick testing and development
- **[Docker Setup](Docker.md)** - Recommended for most users, easiest to set up
- **[Kubernetes Setup](Kubernetes.md)** - For production deployments

After installation, configure your servers using the [Configuration Guide](Configuration.md).
