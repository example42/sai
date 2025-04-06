#!/bin/bash

# Build the application
echo "Building application..."
go build

# Text formatting
GREEN='\033[0;32m'
BLUE='\033[0;34m'
ORANGE='\033[0;33m'
NC='\033[0m' # No Color

# By default sai runs everywhere and offers actions to run on software using normal system commands.
# By default:
# If the action does changes on the systems: it shows the command which is going to run and asks for user confirmation.
# If the action is safe and will not change the system: it shows and run the command

echo -e "${BLUE}===== TESTING PACKAGE MANAGEMENT COMMANDS =====${NC}"

echo -e "${GREEN}Testing install command...${NC}"

# Force unattended installation in interactive mode:
echo -e "${ORANGE}Command: ./sai install nginx -f${NC}"
./sai install nginx -f
echo -e "${ORANGE}Command: ./sai install nginx --force${NC}"
./sai install nginx --force

# Support, via providers, different installation methods
echo -e "${ORANGE}Command: ./sai install nginx --provider source --yes${NC}"
./sai install nginx --provider source --yes
echo -e "${ORANGE}Command: ./sai install nginx --provider container --yes${NC}"
./sai install nginx --provider container --yes
echo -e "${ORANGE}Command: ./sai install nginx --provider snap --yes${NC}"
./sai install nginx --provider snap --yes
echo -e "${ORANGE}Command: ./sai install nginx --provider tarball --yes${NC}"
./sai install nginx --provider tarball --yes
echo -e "${ORANGE}Command: ./sai install nginx --provider upstream --yes${NC}"
./sai install nginx --provider upstream --yes

# Once installed general service management commands:
echo -e "${GREEN}Testing service management commands...${NC}"
# Shows the status of the service nginx for each provider
echo -e "${ORANGE}Command: ./sai status nginx${NC}"
./sai status nginx # Default behaviour: let user select provider, showing options
echo -e "${ORANGE}Command: ./sai status${NC}"
./sai status # Default behvaiours: shows status from default provider

# All the following service management commands are available:
# TO DECIDE: If -y needed here
echo -e "${ORANGE}Command: ./sai status nginx -y${NC}"
./sai status nginx -y
echo -e "${ORANGE}Command: ./sai start nginx -y${NC}"
./sai start nginx -y
echo -e "${ORANGE}Command: ./sai stop nginx -y${NC}"
./sai stop nginx -y
echo -e "${ORANGE}Command: ./sai restart nginx -y${NC}"
./sai restart nginx -y
echo -e "${ORANGE}Command: ./sai reload nginx -y${NC}"
./sai reload nginx -y
echo -e "${ORANGE}Command: ./sai enable nginx -y${NC}"
./sai enable nginx -y
echo -e "${ORANGE}Command: ./sai disable nginx -y${NC}"
./sai disable nginx -y

# All the following package management commands are available:
echo -e "${ORANGE}Command: ./sai install nginx -y${NC}"
./sai install nginx -y
echo -e "${ORANGE}Command: ./sai uninstall nginx -y${NC}"
./sai uninstall nginx -y
echo -e "${ORANGE}Command: ./sai update nginx -y${NC}"
./sai update nginx -y
echo -e "${ORANGE}Command: ./sai upgrade nginx -y${NC}"
./sai upgrade nginx -y

# Various commands to list, search, info, :
echo -e "${ORANGE}Command: ./sai list nginx${NC}"
./sai list nginx
echo -e "${ORANGE}Command: ./sai info nginx${NC}"
./sai info nginx
echo -e "${ORANGE}Command: ./sai log nginx${NC}"
./sai log nginx
echo -e "${ORANGE}Command: ./sai help${NC}"
./sai help
echo -e "${ORANGE}Command: ./sai debug nginx${NC}"
./sai debug nginx
echo -e "${ORANGE}Command: ./sai troubleshoot nginx${NC}"
./sai troubleshoot nginx

# Monitoring commands:
echo -e "${ORANGE}Command: ./sai monitor nginx${NC}"
./sai monitor nginx
echo -e "${ORANGE}Command: ./sai log nginx${NC}"
./sai log nginx
echo -e "${ORANGE}Command: ./sai inspect nginx${NC}"
./sai inspect nginx
echo -e "${ORANGE}Command: ./sai status nginx${NC}"
./sai status nginx
echo -e "${ORANGE}Command: ./sai check nginx${NC}"
./sai check nginx

# Building commands:
echo -e "${ORANGE}Command: ./sai build container -y${NC}"
./sai build container -y
echo -e "${ORANGE}Command: ./sai build rpm -y${NC}"
./sai build rpm -y
echo -e "${ORANGE}Command: ./sai build source -y${NC}"
./sai build source -y

# AI inferences to ask or seek information about the software
echo -e "${ORANGE}Command: ./sai ask${NC}"
./sai ask
echo -e "${ORANGE}Command: ./sai search${NC}"
./sai search

# Manage config files (future)
# ./sai config ....

# Test flags combinations (TO DECIDE which ones to use)
echo -e "${GREEN}Testing flags...${NC}"
echo -e "${ORANGE}Command: ./sai --provider apt install nginx${NC}"
./sai --provider apt install nginx

echo -e "${ORANGE}Command: ./sai install --dry-run${NC}"
./sai install --dry-run

echo -e "${ORANGE}Command: ./sai install apt --yes${NC}"
./sai install apt --yes # as force?
echo -e "${ORANGE}Command: ./sai install apt -y${NC}"
./sai install apt -y

# Some actions should have a default unattended output valid of all software:
echo -e "${ORANGE}Command: ./sai status${NC}"
./sai status # Shows status of all running services
echo -e "${ORANGE}Command: ./sai help${NC}"
./sai help # Shows help of all available commands

echo "All tests completed!"
