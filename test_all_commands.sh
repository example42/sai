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
# By default, the command should explore the available providers support for
# the action desired and show alternative options:
./sai install nginx
# Output like: Available providers for install:
# 1. brew 
# 2. apt
# 3. dnf
# 4. pacman
# User selects provider from the tui

# Alternative to force untattended installation:
./sai install nginx -f
./sai install nginx --force

# Alternative approact software based, is made for unattended operations:
./sai nginx install # Just installs nginx with the default provider

# Multiple installation options (multiple options, which one to choose?
echo -e "${GREEN}Testing differnt install providers...${NC}"
./sai nginx install --provider flatpak
or:
./sai nginx install flatpak
or:
./sai install nginx --provider flatpak
or:
./sai install nginx flatpak 

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
./sai nginx status # Default behvaiours: shows status from default provider

# All the following service management commands are available:
./sai nginx status
./sai nginx start
./sai nginx stop
./sai nginx restart
./sai nginx reload
./sai nginx enable
./sai nginx disable

# All the following package management commands are available:
./sai nginx install
./sai nginx uninstall
./sai nginx update
./sai nginx upgrade

# Various commands to list, search, info, :
./sai nginx list
./sai nginx info
./sai nginx log
./sai nginx help
./sai nginx debug
./sai nginx troubleshoot

# Monitoring commands:
./sai nginx monitor
./sai nginx log
./sai nginx inspect
./sai nginx inspect
./sai nginx status
./sai nginx check

# Building commands:
./sai nginx build container
./sai nginx build rpm
./sai nginx build source 


# AI inferences to ask or seek information about the software
./sai nginx ask
./sai nginx search

# Manage config files (future)
./sai nginx config ....


# Test flags combinations (TO DECIDE which ones to use)
echo -e "${GREEN}Testing flags...${NC}"
./sai --provider apt nginx install

./sai nginx install --dry-run 
./sai nginx install --force

# Force unattended installation
./sai nginx install apt --force
./sai nginx install apt -f
./sai nginx install apt --yes # as force?
./sai nginx install apt -y

# Simulate installation (TO DECIDE)
./sai --provider brew --dry-run nginx install
./sai nginx install --provider apt --noop
./sai nginx install --provider apt --dry-run
./sai nginx install --provider apt --noop
./sai nginx install apt --dry-run
./sai nginx install apt --noop # as dry-run?
./sai nginx install --dry-run
./sai nginx install --provider brew --dry-run

# Some actions should have a default unattended output valid of all software:
./sai status # Shows status of all running services 


echo "All tests completed!"
