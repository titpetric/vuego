# VueGo Playground Docker Setup

This directory contains Docker configuration for the VueGo Playground.

## Quick Start

Run from the repository root:

```bash
# Start the playground with Docker Compose
task docker:up

# Stop the playground
task docker:down

# View logs
task docker:logs

# Open a shell in the container
task docker:shell
```

## Available Tasks

All tasks are accessible from the root directory using the `docker:` namespace:

- `task docker:build` - Build the Docker image
- `task docker:up` - Start the playground container
- `task docker:down` - Stop the playground container
- `task docker:logs` - View container logs
- `task docker:shell` - Open a shell in the running container
- `task docker:clean` - Remove images and containers

## Features

The playground supports:

- **Real-time rendering** - Auto-refresh as you edit templates and data
- **Save functionality** - Save edited templates and data to the filesystem
- **Create new files** - Create new pages (root) or components (components/)
- **Cheatsheet footer** - Quick reference for VueGo syntax
- **Example browser** - Load pre-built examples

## Notes

- When running with embedded filesystem (default), Save and Create buttons are disabled
- When running with a mounted directory, Save and Create are fully functional
- The playground exposes port 8080
- Files can be edited directly in the browser when running with a local filesystem
