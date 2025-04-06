#!/bin/bash

# Build the application
echo "Building application..."
go build

# Text formatting
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# By default sai runs everywhere and offers actions to run on software using normal system commands.
# Bu default it shows the command which is going to run and asks for user confirmation.

echo -e "${BLUE}===== TESTING PACKAGE MANAGEMENT COMMANDS =====${NC}"

echo -e "${GREEN}Testing install command...${NC}"

# Force unattended installation in interactive mode:
./sai install nginx -f
./sai install nginx --force

# Support, via providers, different installation methods
./sai install nginx source
./sai install nginx container
./sai install nginx snap
./sai install nginx tarball
./sai install nginx upstream

# Once installed general service management commands:
echo -e "${GREEN}Testing service management commands...${NC}"
# Shows the status of the service nginx for each provider
./sai status nginx # Default behaviour: let user select provider, showing options
or:
./sai status # Default behvaiours: shows status from default provider

# All the following service management commands are available:
./sai status nginx
./sai start nginx
./sai stop nginx
./sai restart nginx
./sai reload nginx
./sai enable nginx
./sai disable nginx

# All the following package management commands are available:
./sai install nginx
./sai uninstall nginx
./sai update nginx
./sai upgrade nginx

# Various commands to list, search, info, :
./sai list nginx
./sai info nginx
./sai log nginx
./sai help 
./sai debug nginx
./sai troubleshoot nginx

# Monitoring commands:
./sai monitor nginx
./sai log nginx
./sai inspect nginx
./sai status nginx
./sai check nginx

# Building commands:
./sai build container
./sai build rpm
./sai build source 

# AI inferences to ask or seek information about the software
./sai ask
./sai search

# Manage config files (future)
# ./sai config ....


# Test flags combinations (TO DECIDE which ones to use)
echo -e "${GREEN}Testing flags...${NC}"
./sai --provider apt nginx install

./sai install --dry-run 
./sai install --force

# Force unattended installation
./sai install apt --force
./sai install apt -f
./sai install apt --yes # as force?
./sai install apt -y

# Simulate installation (TO DECIDE)
./sai install --provider apt --noop
./sai install --provider apt --dry-run
./sai install --provider apt --noop
./sai install apt --dry-run
./sai install apt --noop # as dry-run?
./sai install --dry-run
./sai install --provider brew --dry-run

# Some actions should have a default unattended output valid of all software:
./sai status # Shows status of all running services 
./sai help # Shows help of all available commands

echo "All tests completed!"
