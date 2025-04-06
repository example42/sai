#!/bin/bash

# Build the application
echo "Building application..."
go build

# Text formatting
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# By default sai runs everywhere and offers actions to run on software using normal system commands.
# By default:
# If the action does changes on the systems: it shows the command which is going to run and asks for user confirmation.
# If the action is safe and will not change the system: it shows and run the command

echo -e "${BLUE}===== TESTING PACKAGE MANAGEMENT COMMANDS =====${NC}"

echo -e "${GREEN}Testing install command...${NC}"

# Force unattended installation in interactive mode:
./sai install nginx -f
./sai install nginx --force

# Support, via providers, different installation methods
./sai install nginx --provider source --yes
./sai install nginx --provider container --yes
./sai install nginx --provider snap --yes
./sai install nginx --provider tarball --yes
./sai install nginx --provider upstream --yes

# Once installed general service management commands:
echo -e "${GREEN}Testing service management commands...${NC}"
# Shows the status of the service nginx for each provider
./sai status nginx # Default behaviour: let user select provider, showing options
or:
./sai status # Default behvaiours: shows status from default provider

# All the following service management commands are available:
# TO DECIDE: If -y needed here
./sai status nginx -y
./sai start nginx -y 
./sai stop nginx -y 
./sai restart nginx -y 
./sai reload nginx -y
./sai enable nginx -y 
./sai disable nginx -y 

# All the following package management commands are available:
./sai install nginx -y
./sai uninstall nginx -y 
./sai update nginx -y 
./sai upgrade nginx -y 

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
./sai build container -y
./sai build rpm -y 
./sai build source -y 

# AI inferences to ask or seek information about the software
./sai ask
./sai search

# Manage config files (future)
# ./sai config ....


# Test flags combinations (TO DECIDE which ones to use)
echo -e "${GREEN}Testing flags...${NC}"
./sai --provider apt nginx install

./sai install --dry-run 

./sai install apt --yes # as force?
./sai install apt -y

# Some actions should have a default unattended output valid of all software:
./sai status # Shows status of all running services 
./sai help # Shows help of all available commands

echo "All tests completed!"
