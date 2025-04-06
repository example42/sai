#!/bin/bash

# Build the application
echo "Building application..."
go build

# Text formatting
GREEN='\033[0;32m'
BLUE='\033[0;34m'
ORANGE='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}===== INTERACTIVE TEST =====${NC}"

echo -e "${GREEN}Testing install command...${NC}"
echo -e "${ORANGE}Command: ./sai install nginx${NC}"
./sai install nginx
# By default, the install command with a software should propmpt the user for confirmation:
# 1. Y/y to proceed and run the shown command to install
# 2. A/a to show all the available install options with the current providers

# To perform unattended installation, user the -y or --yes flag:
echo -e "${ORANGE}Command: ./sai install nginx -y${NC}"
./sai install nginx -y
echo -e "${ORANGE}Command: ./sai install nginx --yes${NC}"
./sai install nginx --yes

echo -e "${GREEN}Testing service management commands...${NC}"
# Shows the status of the service nginx for each provider
echo -e "${ORANGE}Command: ./sai status nginx${NC}"
./sai status nginx # Default behaviour: let user select provider, showing options

# Some actions should have a default unattended output valid of all software:
echo -e "${ORANGE}Command: ./sai status${NC}"
./sai status # Shows status of all running services
echo -e "${ORANGE}Command: ./sai help${NC}"
./sai help # Shows help of all available commands

# Apply special action (applies sai.yaml file with actions to run)
echo -e "${ORANGE}Command: ./sai apply${NC}"
./sai apply # Apply local sai.yaml file in interactive mode
echo -e "${ORANGE}Command: ./sai apply --force${NC}"
./sai apply --force # Apply, for real, sai.yaml file unattended
echo -e "${ORANGE}Command: ./sai apply --dry-run${NC}"
./sai apply --dry-run # Shows what will be done (Default)
echo -e "${ORANGE}Command: ./sai apply examples/devops_station.yaml${NC}"
./sai apply examples/devops_station.yaml # Apply specific sai file (in interactive mode)
echo -e "${ORANGE}Command: ./sai apply examples/devops_station.yaml -y${NC}"
./sai apply examples/devops_station.yaml -y # Apply specific sai file (for real)

