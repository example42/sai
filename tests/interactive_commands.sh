#!/bin/bash

# Build the application
echo "Building application..."
go build

# Text formatting
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}===== INTERACTIVE TEST =====${NC}"

echo -e "${GREEN}Testing install command...${NC}"
# By default, the command should explore the available providers support for
# the action desired and show alternative options:
./sai install nginx
# Output like: Available providers for install:
# 1. brew 
# 2. apt
# 3. dnf
# 4. pacman
# User selects provider from the tui

echo -e "${GREEN}Testing service management commands...${NC}"
# Shows the status of the service nginx for each provider
./sai status nginx # Default behaviour: let user select provider, showing options

# Some actions should have a default unattended output valid of all software:
./sai status # Shows status of all running services 
./sai help # Shows help of all available commands

# Apply special action (applies sai.yaml file with actions to run)
./sai apply # Apply local sai.yaml file in interactive mode
./sai apply --force # Apply, for real, sai.yaml file unattended
./sai apply --dry-run # Shows what will be done (Default)
./sai apply examples/devops_station.yaml # Apply specific sai file (in interactive mode)
./sai apply examples/devops_station.yaml -y # Apply specific sai file (for real)

