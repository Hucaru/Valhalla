# Valhalla Documentation

Welcome to the Valhalla documentation! This directory contains comprehensive guides for setting up and running your MapleStory v28 private server.

## 📖 Documentation Guide

### Getting Started

Start here if you're new to Valhalla:

1. **[Installation Guide](Installation.md)** - Your first stop! Learn how to:
   - Convert Data.wz to Data.nx format
   - Download and set up the MapleStory v28 client
   - Set up the database

### Deployment Options

Choose the deployment method that fits your needs:

- **[Local Setup](Local.md)** - Perfect for development and testing
  - Run directly on your machine
  - Quick to set up
  - Good for learning and experimentation

- **[Docker Setup](Docker.md)** - Recommended for most users
  - Easiest to manage
  - Consistent across all platforms
  - Great for small to medium deployments

- **[Kubernetes Setup](Kubernetes.md)** - For production deployments
  - High availability
  - Easy scaling
  - Production-grade orchestration

### Configuration & Development

- **[Configuration Guide](Configuration.md)** - Complete reference for all server settings
  - Command line flags
  - TOML configuration files
  - Environment variables
  - Performance tuning

- **[Building from Source](Building.md)** - For developers
  - Build instructions
  - Development workflow
  - Cross-compilation
  - Debugging and profiling

## 🗺️ Quick Navigation

| I want to... | Read this guide |
|--------------|----------------|
| Set up Valhalla for the first time | [Installation](Installation.md) → [Local](Local.md) or [Docker](Docker.md) |
| Run a production server | [Installation](Installation.md) → [Kubernetes](Kubernetes.md) |
| Develop and contribute to Valhalla | [Building](Building.md) → [Configuration](Configuration.md) |
| Configure server settings | [Configuration](Configuration.md) |
| Scale to more channels | [Docker](Docker.md#adding-more-channels) or [Kubernetes](Kubernetes.md#scaling-channels) |
| Troubleshoot issues | See troubleshooting sections in each guide |

## 📋 Prerequisites

Before you start, make sure you have:

- ✅ **Data.nx file** - Converted from Data.wz (see [Installation Guide](Installation.md))
- ✅ **MapleStory v28 client** - Download link in [Installation Guide](Installation.md)
- ✅ **MySQL database** - Version 5.7 or later

## 🆘 Getting Help

If you encounter issues:

1. Check the **Troubleshooting** section in the relevant guide
2. Search existing [GitHub Issues](https://github.com/Hucaru/Valhalla/issues)
3. Join our [Discord server](https://discord.gg/KHky9Qy9jF)
4. Create a new [GitHub Issue](https://github.com/Hucaru/Valhalla/issues/new) with:
   - What you were trying to do
   - What happened instead
   - Relevant log output
   - Your deployment method (local/docker/k8s)

## 📝 Documentation Structure

```
docs/
├── README.md           # This file - documentation guide
├── Installation.md     # Data conversion and client setup
├── Local.md           # Running locally
├── Docker.md          # Docker Compose deployment
├── Kubernetes.md      # Kubernetes deployment
├── Configuration.md   # Configuration reference
└── Building.md        # Building from source
```

## 🤝 Contributing

Found an error in the documentation? Want to add more information?

1. Fork the repository
2. Make your changes
3. Submit a pull request

Documentation contributions are always welcome!

## 📚 Additional Resources

- [Main README](../README.md) - Project overview and features
- [GitHub Repository](https://github.com/Hucaru/Valhalla)
- [Discord Community](https://discord.gg/KHky9Qy9jF)
- [NX File Format](https://nxformat.github.io/)
- [go-wztonx-converter](https://github.com/ErwinsExpertise/go-wztonx-converter)
