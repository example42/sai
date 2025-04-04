# SAI (pick the acronym you like)

Do Everything with every Software, Everywhere

## Overview
SAI is a command line that allows you to perfom:
**actions** (many actions) on **every software** (with correct saidata), in different systems (Linux (RedHat, Debian, Suse, Arch, etc.), Windows, macOS, containers, Kubernetes, Nix... ) relying on easy to create **providers** that implement one or more actions.

## Synopsis
```
sai <software> <action> [provider]
sai <action> [software] [provider] # This applies only to reserved actions
```

### Parameters
1. **`<software>`**: The name of the software to manage.
   - Examples: `nginx`, `docker`, `opentofu`, `mysql`, `redis` ... every software for which there's sai data

2. **`<action>`**: The operation to perform on the software.
   - Reserved Actions:
     - `install`, `test`, `build`, `log`, `check`, `observe`, `trace`, `config`, `info`, `debug`, `troubleshoot`, `monitor`, `upgrade`, `uninstall`, `status`, `start`, `stop`, `restart`, `enable`, `disable`, `list`, `search`, `update`,  `ask`, `help`, `apply`... 

3. **`[provider]`** (optional): The specific implementation for software actions.
   - Examples: `rpm`, `apt`, `brew`, `winget`, `helm`, `kubectl`...

## Command line usage examples
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

## As code examples:
All actions can be defined in a sai.yaml file that is applied with  `sia apply`:
```yaml
sai:
  install:
    - vscode
    - git
    - docker
    - docker-compose
    - awscli
    - gcloud
    - azure-cli
    - kubectl
    - helm
    - terraform
```

Find more examples of sai commands as code in the [examples](examples/) directory..

## Features
- **Cross-Platform Support**: Works seamlessly across Linux, macOS, Windows, and containerized environments.
- **Provider Abstraction**: Handles provider-specific commands internally for simplicity.
- **Providers simple**: New providers can be added via just yaml files with the commands to run for supported actions
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
   go build -o sai
   ```

3. Run the CLI:
   ```
   # Interactive mode:
   ./sai <action> [software]

   # Unattended mode:
   ./sai <software> <action> [provider]

   ```

## Contributing
Contributions are welcome! Please submit issues or pull requests to improve SAI.