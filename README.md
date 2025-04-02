# SAI (Find an achronism that you like)

## Human invented
## AI driven, designed, and supported

## Overview
SAI is a solution designed to handle various software-related use cases across different environments. It simplifies tasks such as installing, managing, configuring, monitoring, testing, debugging, and troubleshooting software on platforms like Linux (RedHat, Debian, Suse, Arch, etc.), Windows, macOS, containers, and Kubernetes.

## Synopsis
```
sai <software> <action> [provider]
```

### Parameters
1. **`<software>`**: The name of the software to manage.
   - Examples: `nginx`, `docker`, `opentofu`, `mysql`, `redis` ... every software for which there's sai data

2. **`<action>`**: The operation to perform on the software.
   - Supported Actions: []
   - Supported Actions [TODO]:
     - `install`, `test`, `build`, `log`, `check`, `observe`, `trace`, `config`, `info`, `debug`, `troubleshoot`, `monitor`, `upgrade`, `uninstall`, `status`, `start`, `stop`, `restart`, `enable`, `disable`, `list`, `search`, `update`,  `ask`, `help`... 

3. **`[provider]`** (optional): The specific implementation for software actions.
   - Examples: `rpm`, `apt`, `brew`, `winget`, `helm`, `kubectl`...

## Examples
1. Install an application and manage it:
   ```
   sai nginx install
   sai nginx status
   sai nginx start
   sai nginx stop
   sai nginx enable
   sai nginx disable
   ```

2. Check, monitor and troubleshoot:
   ```
   sai tomcat check
   sai tomcat log
   sai tomcat troubleshoot
   sai tomcat debug
   sai tomcat info
   sai tomcat monitor
   sai tomcat observe
   sai tomcat trace
   ```

3. Build different images:
   ```
   sai myapp build rpm
   sai myapp build apt
   sai myapp build brew
   sai myapp build winget
   sai myapp build helm
   sai myapp build container   
   ```

4. Ask information about a software:
   ```
   sai terraform ask
   sai terraform help
   sai terraform info
   sai terraform config
   sai terraform troubleshoot
   ```

## Features
- **Cross-Platform Support**: Works seamlessly across Linux, macOS, Windows, and containerized environments.
- **Provider Abstraction**: Handles provider-specific commands internally for simplicity.
- **Extensibility**: Easily add new actions, software, and providers.
- **Error Handling**: Provides meaningful error messages for unsupported actions, software, or providers.
- **AI Generated Software data**: The saidata for supported software is generated by AI using custom fine tuned models
- **Automatically Tested**: Saidata and implementation is automaticaally tested using a custom test framework
  
## Getting Started
1. Clone the repository:
   ```
   git clone https://github.com/example42/sai.git
   ```
   cd sai
   ```

2. Build the project:
   ```
   go build -o sai ./cmd/main.go
   ```

3. Run the CLI:
   ```
   ./sai <software> <action> [provider]
   ```

## Contributing
Contributions are welcome! Please submit issues or pull requests to improve SAI.