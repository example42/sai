# Requirements Document

## Introduction

The SAI CLI application is a lightweight command-line tool for executing software management actions using provider-based configurations. The tool aims to provide a unified interface for managing software across different operating systems and package managers with the motto "Do everything on every software on every OS". The application will support hierarchical saidata structures, automatic provider detection, and cross-platform compatibility.

## Requirements

### Requirement 1

**User Story:** As a system administrator, I want to install software packages using a unified command interface, so that I can manage software consistently across different operating systems and package managers.

#### Acceptance Criteria

1. WHEN the user runs `sai install <software>` THEN the system SHALL detect the appropriate provider and install the specified software
2. WHEN the user specifies a provider with `--provider/-p <name>` THEN the system SHALL use that specific provider instead of auto-detection
3. WHEN multiple package_manager providers are available THEN the system SHALL display a list of possible choices showing provider-specific package name and version for that software AND prompt the user for which provider to use
4. WHEN the `--yes/-y` flag is provided THEN the system SHALL perform unattended installation using the higher priority provider which can provide the software
5. WHEN the `--dry-run` flag is used THEN the system SHALL show what would be executed without actually running commands

### Requirement 2

**User Story:** As a system administrator, I want to manage software lifecycle operations (uninstall, upgrade, search, info, version), so that I can maintain software packages effectively.

#### Acceptance Criteria

1. WHEN the user runs `sai uninstall <software>` THEN the system SHALL remove the specified software package AND IF software is installed using different providers THEN the system SHALL provide a list for user selection
2. WHEN the user runs `sai upgrade <software>` THEN the system SHALL update the software to its latest version
3. WHEN the user runs `sai search <software>` THEN the system SHALL search for available packages in repositories across all available providers AND execute without requiring further user confirmation
4. WHEN the user runs `sai info <software>` THEN the system SHALL display detailed information about the software AND execute without requiring further user confirmation
5. WHEN the user runs `sai version <software>` THEN the system SHALL show the software version for all providers which provide it AND highlight whether the package is already installed

### Requirement 3

**User Story:** As a system administrator, I want to control software services (start, stop, restart, enable, disable, status), so that I can manage running services effectively.

#### Acceptance Criteria

1. WHEN the user runs `sai start <software>` THEN the system SHALL start the specified software service
2. WHEN the user runs `sai stop <software>` THEN the system SHALL stop the running software service
3. WHEN the user runs `sai restart <software>` THEN the system SHALL restart the software service
4. WHEN the user runs `sai enable <software>` THEN the system SHALL enable automatic startup at boot
5. WHEN the user runs `sai disable <software>` THEN the system SHALL disable automatic startup at boot
6. WHEN the user runs `sai status <software>` THEN the system SHALL check and display the current service status AND execute without requiring further user confirmation

### Requirement 4

**User Story:** As a system administrator, I want to monitor software performance and view logs, so that I can troubleshoot and optimize system performance.

#### Acceptance Criteria

1. WHEN the user runs `sai logs <software>` THEN the system SHALL display service logs and output
2. WHEN the user runs `sai config <software>` THEN the system SHALL show configuration files
3. WHEN the user runs `sai check <software>` THEN the system SHALL verify if software is working correctly
4. WHEN the user runs `sai cpu <software>` THEN the system SHALL display CPU usage statistics for the software
5. WHEN the user runs `sai memory <software>` THEN the system SHALL show memory usage statistics for the software
6. WHEN the user runs `sai io <software>` THEN the system SHALL display I/O statistics for the software

### Requirement 5

**User Story:** As a system administrator, I want to view general system information and installed software, so that I can get an overview of the system state.

#### Acceptance Criteria

1. WHEN the user runs `sai list` THEN the system SHALL display all installed software packages for all the available providers
2. WHEN the user runs `sai logs` without software parameter THEN the system SHALL display general system service logs
3. WHEN the user runs `sai cpu` without software parameter THEN the system SHALL display general CPU usage statistics
4. WHEN the user runs `sai memory` without software parameter THEN the system SHALL display general memory usage statistics
5. WHEN the user runs `sai io` without software parameter THEN the system SHALL display general I/O statistics


### Requirement 6

**User Story:** As a system administrator, I want to execute batch operations and view system statistics, so that I can automate tasks and understand system capabilities.

#### Acceptance Criteria

1. WHEN the user runs `sai apply <action_file>` THEN the system SHALL execute multiple software management actions from YAML/JSON file
2. WHEN the user runs `sai stats` THEN the system SHALL display comprehensive statistics about available providers, actions, and system capabilities
3. WHEN the user runs `sai saidata` THEN the system SHALL manage saidata repository operations including updates and synchronization
4. WHEN the apply file format is invalid THEN the system SHALL validate against the applydata-0.1-schema.json schema and report detailed validation errors

### Requirement 7

**User Story:** As a system administrator, I want to use global options with any command, so that I can customize the behavior and output format of the tool.

#### Acceptance Criteria

1. WHEN the user uses `--config/-c <path>` THEN the system SHALL use the specified custom configuration file
2. WHEN the user uses `--verbose/-v` THEN the system SHALL enable detailed output and logging information
3. WHEN the user uses `--dry-run` THEN the system SHALL show what would be executed without running commands
4. WHEN the user uses `--yes/-y` THEN the system SHALL automatically confirm all prompts
5. WHEN the user uses `--quiet/-q` THEN the system SHALL suppress non-essential output
6. WHEN the user uses `--json` THEN the system SHALL output results in JSON format

### Requirement 8

**User Story:** As a system administrator, I want the tool to work across different platforms and automatically detect providers, so that I can use the same interface regardless of the operating system.

#### Acceptance Criteria

1. WHEN the tool runs on Linux, macOS, or Windows THEN the system SHALL function correctly on all platforms
2. WHEN the tool needs to determine a provider THEN the system SHALL automatically detect and prioritize appropriate providers based on the executable key in the provider data and platform compatibility
3. WHEN saidata is needed THEN the system SHALL support hierarchical saidata structure following the pattern software/{prefix}/{software}/default.yaml with OS-specific overrides
4. WHEN saidata repository is accessed THEN the system SHALL use Git-based management with zip fallback for offline scenarios
5. WHEN software repositories are defined in saidata THEN the system SHALL automatically manage repository setup and configuration

### Requirement 9

**User Story:** As a system administrator, I want appropriate confirmation prompts for system-changing operations, so that I can prevent accidental modifications while allowing unattended information queries.

#### Acceptance Criteria

1. WHEN an action changes the system state THEN the system SHALL require user confirmation by default
2. WHEN an action only displays information THEN the system SHALL execute without confirmation prompts
3. WHEN the `--yes/-y` flag is provided THEN the system SHALL skip confirmation prompts for all operations
4. WHEN the `--quiet/-q` flag is used THEN the system SHALL minimize output while maintaining essential confirmations

### Requirement 10

**User Story:** As a system administrator, I want coherent and clear output and exit code from every action

#### Acceptance Criteria

1. WHEN a command is going to be executed THEN the system SHALL display it in bold text before command execution and before any user prompts
2. WHEN multiple providers are used THEN the system SHALL display the provider name with a configurable background color
3. WHEN a command is executed THEN the system SHALL show the output in normal text followed by highlighted exit status (Green for success, Red for failure)
4. WHEN any executed command fails THEN the sai command exit code SHALL be 1 AND WHEN all commands succeed THEN the exit code SHALL be 0
5. WHEN validating commands THEN the system SHALL only show and execute commands that can be executed AND SHALL NOT attempt to run commands with unavailable providers or on non-existent software

### Requirement 11

**User Story:** As a system administrator, I want automatic saidata repository management, so that I can use SAI without manual setup and always have up-to-date software definitions.

#### Acceptance Criteria

1. WHEN SAI is executed for the first time THEN the system SHALL display a relevant message about downloading saidata repository
2. WHEN running as a normal user THEN the system SHALL download saidata to $HOME/.sai/saidata directory
3. WHEN running as root THEN the system SHALL download saidata to /etc/sai/saidata directory (or Windows equivalent)
4. WHEN downloading saidata THEN the system SHALL use Git clone as the primary method with zip download as fallback
5. WHEN the `sai saidata` command is executed THEN the system SHALL provide saidata management operations including update, status, and synchronization

### Requirement 12

**User Story:** As a developer, I want detailed debug information when troubleshooting SAI issues, so that I can diagnose problems effectively.

#### Acceptance Criteria

1. WHEN the `--debug` flag is provided THEN the system SHALL display detailed debug information including provider detection, template resolution, and command execution details
2. WHEN debug mode is enabled THEN the system SHALL show internal state information, configuration loading, and decision-making processes
3. WHEN debug mode is enabled THEN the system SHALL log all template variable resolutions and provider selection logic
4. WHEN debug mode is enabled THEN the system SHALL display timing information for operations and performance metrics

### Requirement 13

**User Story:** As a system administrator, I want accurate provider detection and system statistics, so that I can understand which providers are actually available on my system.

#### Acceptance Criteria

1. WHEN `sai stats` is executed THEN the system SHALL only show providers that are actually available based on executable detection
2. WHEN a provider defines an `executable` field THEN the system SHALL use that executable to determine provider availability
3. WHEN provider detection fails THEN the system SHALL not list that provider as available in stats output
4. WHEN multiple providers are available THEN the system SHALL show accurate availability status for each provider

### Requirement 14

**User Story:** As a system administrator, I want to check software versions using SAI's version command, so that I can see installed software versions across different providers.

#### Acceptance Criteria

1. WHEN `sai version <software>` is executed THEN the system SHALL show the version of the software if installed for all available providers
2. WHEN a provider has a version action defined THEN the system SHALL use that action's command to retrieve version information
3. WHEN `sai --version` is executed THEN the system SHALL show SAI's own version information
4. WHEN software is not installed THEN the system SHALL indicate that the software is not installed for that provider
5. WHEN version information cannot be retrieved THEN the system SHALL show an appropriate error message

### Requirement 15

**User Story:** As a system administrator, I want compact and informative output when multiple providers are available, so that I can quickly understand my options and make informed decisions.

#### Acceptance Criteria

1. WHEN multiple providers are available for an action THEN the system SHALL display provider name and the full command that will be executed
2. WHEN showing provider options THEN the system SHALL display "Command: (The full command executed/to execute to get the output)" instead of package/version/description
3. WHEN an action only displays information and doesn't modify the system THEN the system SHALL execute the command for each available provider without asking for selection
4. WHEN an action modifies the system THEN the system SHALL still prompt for provider selection
5. WHEN displaying provider options THEN the system SHALL show availability status for each provider

### Requirement 16

**User Story:** As a system administrator, I want all SAI actions to work correctly, so that I can actually manage software using the tool.

#### Acceptance Criteria

1. WHEN any SAI command is executed THEN the system SHALL successfully complete the requested action without errors
2. WHEN template resolution occurs THEN all saidata template functions SHALL resolve correctly to actual values
3. WHEN provider actions are executed THEN the commands SHALL be properly constructed and executed
4. WHEN saidata is loaded THEN the system SHALL correctly parse and use the software definitions
5. WHEN providers are loaded THEN the system SHALL correctly parse and execute provider actions
6. WHEN any action fails THEN the system SHALL provide clear error messages indicating what went wrong and how to fix it

### Requirement 17

**User Story:** As a system administrator, I want consistent naming for packages and services in saidata and providers, so that template functions work predictably across all configurations.

#### Acceptance Criteria

1. WHEN defining packages in saidata THEN the system SHALL use `package_name` parameter instead of `name` for package identification
2. WHEN using template functions THEN `sai_package()` SHALL reference the `package_name` field consistently
3. WHEN using template functions THEN `sai_service()` SHALL continue to reference the `service_name` field as before
4. WHEN updating providers THEN all existing provider templates SHALL be updated to use `package_name` instead of `name`
5. WHEN updating schemas THEN the saidata schema SHALL be updated to require `package_name` field for packages