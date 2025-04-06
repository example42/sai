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
./sai install nginx
# By default, the install command with a software should propmpt the user for confirmation:
# 1. Y/y to proceed and run the shown command to install 
# 2. A/a to show all the available install options with the current providers

# To perform unattended installation, user the -y or --yes flag:
./sai install nginx -y
./sai install nginx --yes

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

