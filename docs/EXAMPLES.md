# SAI CLI Examples

This document provides comprehensive examples of using SAI CLI with real-world software configurations. All examples use the existing saidata samples included with SAI.

## Table of Contents

- [Basic Usage](#basic-usage)
- [Web Servers](#web-servers)
- [Databases](#databases)
- [Container Platforms](#container-platforms)
- [Monitoring and Observability](#monitoring-and-observability)
- [Development Tools](#development-tools)
- [Batch Operations](#batch-operations)
- [Advanced Scenarios](#advanced-scenarios)

## Basic Usage

### Getting Started

```bash
# Check SAI version
sai version

# List all available providers
sai stats

# Show help for any command
sai install --help
```

### Software Installation

```bash
# Install with automatic provider detection
sai install nginx

# Install with specific provider
sai install nginx --provider apt

# Install with confirmation bypass
sai install nginx --yes

# See what would be executed without running
sai install nginx --dry-run

# Install with verbose output
sai install nginx --verbose
```

## Web Servers

### Apache HTTP Server

Apache is available in the saidata samples at `docs/saidata_samples/ap/apache/`.

```bash
# Install Apache
sai install apache

# Start Apache service
sai start apache

# Enable Apache to start at boot
sai enable apache

# Check Apache status
sai status apache

# View Apache configuration
sai config apache

# Check Apache logs
sai logs apache

# Monitor Apache performance
sai cpu apache
sai memory apache
sai io apache

# Test if Apache is responding
sai check apache

# Get Apache version information
sai version apache

# Stop Apache
sai stop apache

# Restart Apache (useful after config changes)
sai restart apache

# Uninstall Apache
sai uninstall apache
```

#### Platform-Specific Apache Examples

```bash
# On Ubuntu 22.04 (uses apache2 package)
sai install apache --provider apt

# On CentOS 8 (uses httpd package)
sai install apache --provider dnf

# On macOS (uses httpd via Homebrew)
sai install apache --provider brew
```

### Nginx

```bash
# Complete Nginx setup workflow
sai install nginx
sai start nginx
sai enable nginx
sai status nginx

# View Nginx configuration files
sai config nginx

# Monitor Nginx in real-time
sai logs nginx --follow

# Performance monitoring
sai cpu nginx
sai memory nginx

# Health check
sai check nginx

# Reload configuration (restart)
sai restart nginx
```

## Databases

### MySQL

MySQL configuration is available in `docs/saidata_samples/my/mysql/`.

```bash
# Install MySQL
sai install mysql

# Start MySQL service
sai start mysql

# Enable MySQL at boot
sai enable mysql

# Check MySQL status
sai status mysql

# View MySQL configuration
sai config mysql

# Monitor MySQL logs
sai logs mysql

# Performance monitoring
sai cpu mysql
sai memory mysql
sai io mysql

# Check MySQL connectivity
sai check mysql

# Get MySQL version
sai version mysql
```

### MongoDB

MongoDB configuration is in `docs/saidata_samples/mo/mongodb/`.

```bash
# Install MongoDB
sai install mongodb

# Start MongoDB
sai start mongodb

# Check if MongoDB is running
sai status mongodb

# View MongoDB logs
sai logs mongodb

# Monitor MongoDB performance
sai cpu mongodb
sai memory mongodb

# Check MongoDB health
sai check mongodb
```

### Redis

Redis configuration is in `docs/saidata_samples/re/redis/`.

```bash
# Install and start Redis
sai install redis
sai start redis
sai enable redis

# Check Redis status
sai status redis

# Monitor Redis
sai logs redis
sai memory redis

# Test Redis connectivity
sai check redis
```

## Container Platforms

### Docker

Docker configuration is in `docs/saidata_samples/do/docker/`.

```bash
# Install Docker
sai install docker

# Start Docker service
sai start docker

# Enable Docker at boot
sai enable docker

# Check Docker status
sai status docker

# View Docker logs
sai logs docker

# Monitor Docker daemon
sai cpu docker
sai memory docker

# Verify Docker installation
sai check docker

# Get Docker version
sai version docker
```

### Kubernetes

Kubernetes configuration is in `docs/saidata_samples/ku/kubernetes/`.

```bash
# Install Kubernetes tools
sai install kubernetes

# Check Kubernetes status
sai status kubernetes

# View Kubernetes logs
sai logs kubernetes

# Monitor Kubernetes components
sai cpu kubernetes
sai memory kubernetes
```

## Monitoring and Observability

### Elasticsearch

Elasticsearch configuration is in `docs/saidata_samples/el/elasticsearch/`.

```bash
# Install Elasticsearch
sai install elasticsearch

# Start Elasticsearch
sai start elasticsearch

# Enable at boot
sai enable elasticsearch

# Check cluster status
sai status elasticsearch

# Monitor Elasticsearch logs
sai logs elasticsearch

# Performance monitoring
sai cpu elasticsearch
sai memory elasticsearch
sai io elasticsearch

# Health check
sai check elasticsearch

# View Elasticsearch configuration
sai config elasticsearch
```

### Grafana

Grafana configuration is in `docs/saidata_samples/gr/grafana/`.

```bash
# Install and setup Grafana
sai install grafana
sai start grafana
sai enable grafana

# Check Grafana status
sai status grafana

# View Grafana logs
sai logs grafana

# Monitor Grafana performance
sai cpu grafana
sai memory grafana

# Access Grafana configuration
sai config grafana
```

### Prometheus

Prometheus configuration is in `docs/saidata_samples/pr/prometheus/`.

```bash
# Install Prometheus
sai install prometheus

# Start Prometheus
sai start prometheus

# Check Prometheus status
sai status prometheus

# View Prometheus configuration
sai config prometheus

# Monitor Prometheus logs
sai logs prometheus

# Performance monitoring
sai cpu prometheus
sai memory prometheus
```

## Development Tools

### Jenkins

Jenkins configuration is in `docs/saidata_samples/je/jenkins/`.

```bash
# Install Jenkins
sai install jenkins

# Start Jenkins
sai start jenkins

# Enable Jenkins at boot
sai enable jenkins

# Check Jenkins status
sai status jenkins

# View Jenkins logs
sai logs jenkins

# Monitor Jenkins performance
sai cpu jenkins
sai memory jenkins

# Access Jenkins configuration
sai config jenkins

# Health check
sai check jenkins
```

### Terraform

Terraform configuration is in `docs/saidata_samples/te/terraform/`.

```bash
# Install Terraform
sai install terraform

# Check Terraform version
sai version terraform

# Terraform doesn't run as a service, so these commands show system info
sai cpu terraform    # Shows system CPU usage
sai memory terraform # Shows system memory usage
```

## Batch Operations

### Web Server Stack

Create a `web-stack.yaml` file:

```yaml
version: "0.1"
description: "Complete web server stack setup"
actions:
  - action: install
    software: nginx
    provider: apt
  - action: install
    software: mysql
    provider: apt
  - action: start
    software: nginx
  - action: start
    software: mysql
  - action: enable
    software: nginx
  - action: enable
    software: mysql
```

Execute the batch operation:

```bash
sai apply web-stack.yaml
```

### Monitoring Stack

Create a `monitoring-stack.yaml` file:

```yaml
version: "0.1"
description: "Complete monitoring stack"
actions:
  - action: install
    software: prometheus
  - action: install
    software: grafana
  - action: install
    software: elasticsearch
  - action: start
    software: prometheus
  - action: start
    software: grafana
  - action: start
    software: elasticsearch
  - action: enable
    software: prometheus
  - action: enable
    software: grafana
  - action: enable
    software: elasticsearch
```

Execute:

```bash
sai apply monitoring-stack.yaml --yes
```

### Development Environment

Create a `dev-env.yaml` file:

```yaml
version: "0.1"
description: "Development environment setup"
actions:
  - action: install
    software: docker
  - action: install
    software: jenkins
  - action: start
    software: docker
  - action: start
    software: jenkins
  - action: enable
    software: docker
  - action: enable
    software: jenkins
```

Execute:

```bash
sai apply dev-env.yaml
```

## Advanced Scenarios

### Cross-Platform Installation

```bash
# Install on different platforms with automatic detection
# On Ubuntu
sai install apache  # Uses apt, installs apache2

# On CentOS
sai install apache  # Uses dnf/yum, installs httpd

# On macOS
sai install apache  # Uses brew, installs httpd

# Force specific provider regardless of platform
sai install apache --provider docker  # Uses Docker on any platform
```

### Provider Selection

```bash
# When multiple providers are available, SAI will prompt
sai install nginx
# Output:
# Multiple providers available for nginx:
# 1. apt - Package: nginx, Version: 1.18.0, Status: Available
# 2. docker - Package: nginx:latest, Version: latest, Status: Available
# 3. snap - Package: nginx, Version: 1.18.0, Status: Available
# Select provider [1-3]:

# Skip prompt with --yes (uses highest priority provider)
sai install nginx --yes

# Force specific provider
sai install nginx --provider docker
```

### Monitoring Multiple Services

```bash
# Monitor all services in a stack
for service in nginx mysql redis; do
  echo "=== $service ==="
  sai status $service
  sai cpu $service
  sai memory $service
  echo
done
```

### Configuration Management

```bash
# View configuration files for multiple services
sai config apache
sai config nginx
sai config mysql

# Check logs for troubleshooting
sai logs apache --tail 50
sai logs nginx --tail 50
sai logs mysql --tail 50
```

### Health Monitoring

```bash
# Create a health check script
#!/bin/bash
services=("nginx" "apache" "mysql" "redis" "elasticsearch")

for service in "${services[@]}"; do
  echo -n "Checking $service: "
  if sai check $service >/dev/null 2>&1; then
    echo "✓ Healthy"
  else
    echo "✗ Unhealthy"
    sai logs $service --tail 10
  fi
done
```

### Performance Monitoring

```bash
# Monitor resource usage across services
echo "Service Resource Usage:"
echo "======================"
printf "%-15s %-10s %-10s %-10s\n" "Service" "CPU%" "Memory%" "I/O"
printf "%-15s %-10s %-10s %-10s\n" "-------" "----" "-------" "---"

for service in nginx apache mysql redis; do
  cpu=$(sai cpu $service --json | jq -r '.cpu_percent // "N/A"')
  mem=$(sai memory $service --json | jq -r '.memory_percent // "N/A"')
  io=$(sai io $service --json | jq -r '.io_rate // "N/A"')
  printf "%-15s %-10s %-10s %-10s\n" "$service" "$cpu" "$mem" "$io"
done
```

### Automated Deployment

```bash
#!/bin/bash
# Automated deployment script

set -e

echo "Starting deployment..."

# Install required software
sai install nginx --yes
sai install mysql --yes
sai install redis --yes

# Start services
sai start nginx
sai start mysql
sai start redis

# Enable at boot
sai enable nginx
sai enable mysql
sai enable redis

# Verify deployment
echo "Verifying deployment..."
sai check nginx
sai check mysql
sai check redis

echo "Deployment completed successfully!"

# Show status
sai status nginx
sai status mysql
sai status redis
```

### Disaster Recovery

```bash
#!/bin/bash
# Service recovery script

services=("nginx" "apache" "mysql" "redis")

for service in "${services[@]}"; do
  echo "Checking $service..."
  
  if ! sai status $service | grep -q "active"; then
    echo "Restarting $service..."
    sai restart $service
    
    # Wait and verify
    sleep 5
    if sai check $service; then
      echo "$service recovered successfully"
    else
      echo "Failed to recover $service"
      sai logs $service --tail 20
    fi
  else
    echo "$service is running normally"
  fi
done
```

## JSON Output and Automation

### Using JSON Output

```bash
# Get structured output for automation
sai list --json | jq '.installed[] | select(.name == "nginx")'

# Monitor services with JSON
sai status nginx --json | jq '.status'

# Get performance metrics
sai cpu nginx --json | jq '.cpu_percent'
sai memory nginx --json | jq '.memory_usage'
```

### Integration with Monitoring Systems

```bash
#!/bin/bash
# Send metrics to monitoring system

services=("nginx" "apache" "mysql")

for service in "${services[@]}"; do
  cpu=$(sai cpu $service --json | jq -r '.cpu_percent')
  memory=$(sai memory $service --json | jq -r '.memory_percent')
  
  # Send to monitoring system (example)
  curl -X POST "http://monitoring.example.com/metrics" \
    -H "Content-Type: application/json" \
    -d "{\"service\":\"$service\",\"cpu\":$cpu,\"memory\":$memory}"
done
```

## Troubleshooting Examples

### Common Issues

```bash
# Service won't start
sai start nginx
sai logs nginx --tail 50  # Check recent logs
sai config nginx         # Verify configuration

# High resource usage
sai cpu nginx
sai memory nginx
sai io nginx

# Service health issues
sai check nginx
sai status nginx --verbose
```

### Debug Mode

```bash
# Enable verbose output for troubleshooting
sai install nginx --verbose --dry-run

# Check what commands would be executed
sai start nginx --dry-run

# Verify provider availability
sai stats --provider apt
```

---

These examples demonstrate the power and flexibility of SAI CLI across different software types and use cases. The consistent interface makes it easy to manage diverse software stacks with a single tool.

For more information, see the [main documentation](../README.md) or the [Provider Development Guide](PROVIDER_DEVELOPMENT.md).