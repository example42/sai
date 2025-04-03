#!/bin/bash

# Build the application
echo "Building application..."
go build

# Text formatting
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}===== TESTING PACKAGE MANAGEMENT COMMANDS =====${NC}"

# Test package management commands with nginx
echo -e "${GREEN}Testing install command...${NC}"
./sai nginx install

echo -e "${GREEN}Testing status command...${NC}"
./sai nginx status

echo -e "${GREEN}Testing list command...${NC}"
./sai nginx list

echo -e "${GREEN}Testing search command...${NC}"
./sai nginx search

echo -e "${GREEN}Testing info command...${NC}"
./sai nginx info

echo -e "${GREEN}Testing upgrade command...${NC}"
./sai nginx upgrade

echo -e "${GREEN}Testing uninstall command...${NC}"
./sai nginx uninstall

echo -e "${BLUE}===== TESTING SERVICE MANAGEMENT COMMANDS =====${NC}"

# Test service management commands with redis
echo -e "${GREEN}Testing start command...${NC}"
./sai redis start

echo -e "${GREEN}Testing stop command...${NC}"
./sai redis stop

echo -e "${GREEN}Testing restart command...${NC}"
./sai redis restart

echo -e "${GREEN}Testing enable command...${NC}"
./sai redis enable

echo -e "${GREEN}Testing disable command...${NC}"
./sai redis disable

echo -e "${BLUE}===== TESTING WITH FLAGS BEFORE ARGUMENTS =====${NC}"

# Test with flags before arguments
echo -e "${GREEN}Testing provider flag before arguments...${NC}"
./sai --provider apt nginx install

echo -e "${GREEN}Testing dry-run flag before arguments...${NC}"
./sai --dry-run nginx install

echo -e "${GREEN}Testing multiple flags before arguments...${NC}"
./sai --provider brew --dry-run nginx install

echo -e "${BLUE}===== TESTING WITH FLAGS AFTER ARGUMENTS =====${NC}"

# Test with flags after arguments
echo -e "${GREEN}Testing provider flag after arguments...${NC}"
./sai nginx install --provider apt

echo -e "${GREEN}Testing dry-run flag after arguments...${NC}"
./sai nginx install --dry-run

echo -e "${GREEN}Testing multiple flags after arguments...${NC}"
./sai nginx install --provider brew --dry-run

echo -e "${BLUE}===== TESTING OTHER COMMANDS =====${NC}"

# Test other commands
echo -e "${GREEN}Testing help command...${NC}"
./sai nginx help

echo -e "${GREEN}Testing debug command...${NC}"
./sai nginx debug

echo -e "${GREEN}Testing troubleshoot command...${NC}"
./sai nginx troubleshoot

echo -e "${GREEN}Testing config command...${NC}"
./sai nginx config

echo "All tests completed!"
