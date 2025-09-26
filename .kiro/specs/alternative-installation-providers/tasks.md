# Implementation Plan

## Task Overview

Convert the alternative installation providers design into a series of coding tasks for implementing source, binary, and script providers with complete template function support and seamless SAI integration.

## Implementation Tasks

### Phase 1: Type System and Schema Foundation

- [x] 1. Extend SoftwareData type with alternative installation support
  - Add Sources, Binaries, and Scripts fields to SoftwareData struct
  - Add corresponding fields to ProviderConfig struct
  - Update JSON marshaling/unmarshaling methods
  - Add helper methods for accessing sources, binaries, and scripts by name/index
  - _Requirements: 5.4, 5.5, 5.7_

- [x] 1.1 Define Source type and supporting structures
  - Create Source struct with all required fields (name, url, build_system, etc.)
  - Create SourceCustomCommands struct for build step overrides
  - Add validation methods for build system types and required fields
  - Implement default value generation for build directories and install prefixes
  - _Requirements: 1.7, 5.1, 5.6_

- [x] 1.2 Define Binary type and supporting structures
  - Create Binary struct with download and installation configuration
  - Create ArchiveConfig struct for handling compressed downloads
  - Create BinaryCustomCommands struct for installation step overrides
  - Add OS/architecture templating support for download URLs
  - _Requirements: 2.5, 2.8, 5.2_

- [x] 1.3 Define Script type and supporting structures
  - Create Script struct with execution configuration
  - Create ScriptCustomCommands struct for execution step overrides
  - Add environment variable and argument handling
  - Implement timeout and working directory configuration
  - _Requirements: 3.3, 3.7, 5.3_

### Phase 2: Template Function Implementation

- [x] 2. Implement sai_source template function
  - Add sai_source function to template engine function map
  - Implement field resolution with provider override support
  - Add default value generation for missing fields (build_dir, source_dir, etc.)
  - Handle build system-specific command generation (autotools, cmake, make, etc.)
  - Add comprehensive error handling and validation
  - _Requirements: 4.1, 4.5, 4.6, 1.8_

- [x] 2.1 Implement sai_binary template function
  - Add sai_binary function to template engine function map
  - Implement OS/architecture placeholder resolution in URLs
  - Add checksum verification and archive extraction logic
  - Handle installation path and permission configuration
  - Add binary-specific validation and error handling
  - _Requirements: 4.2, 2.2, 2.4, 2.6_

- [x] 2.2 Implement sai_script template function
  - Add sai_script function to template engine function map
  - Implement environment variable and argument resolution
  - Add script interpreter detection and configuration
  - Handle working directory and timeout settings
  - Add script-specific validation and security checks
  - _Requirements: 4.3, 3.4, 3.8, 3.2_

- [x] 2.3 Add template function error handling and validation
  - Implement graceful degradation when template functions fail to resolve
  - Add detailed error messages for missing configuration fields
  - Update template validation to include new function signatures
  - Add template function testing framework for all new functions
  - _Requirements: 4.4, 4.8, 5.6_

### Phase 3: Provider Implementation

- [x] 3. Complete source provider implementation
  - Update existing source.yaml provider with comprehensive action definitions
  - Implement multi-step build workflow (prerequisites, download, extract, configure, build, install)
  - Add build system detection and appropriate command generation
  - Implement rollback procedures for failed builds
  - Add source build validation and version detection
  - _Requirements: 1.1, 1.2, 1.4, 1.5, 1.7_

- [x] 3.1 Create binary provider implementation
  - Create providers/binary.yaml with complete action definitions
  - Implement download workflow with checksum verification
  - Add archive extraction support for multiple formats (tar.gz, zip, etc.)
  - Implement binary installation with correct permissions and PATH placement
  - Add binary-specific uninstall and upgrade procedures
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.6, 2.7_

- [x] 3.2 Create script provider implementation
  - Create providers/script.yaml with complete action definitions
  - Implement script download and checksum verification
  - Add secure script execution with environment variable support
  - Implement automatic confirmation handling for interactive scripts
  - Add script-specific rollback and uninstall procedures
  - _Requirements: 3.1, 3.2, 3.4, 3.5, 3.6, 3.8_

### Phase 4: Integration and Compatibility

- [x] 4. Update schema validation and documentation
  - Update saidata-0.2-schema.json with new type definitions
  - Add comprehensive field validation for all new types
  - Update schema documentation with examples and field descriptions
  - Validate schema compatibility with existing saidata files
  - _Requirements: 5.1, 5.2, 5.3, 5.6, 5.8_

- [x] 4.1 Integrate with existing SAI features
  - Ensure alternative providers work with standard SAI actions (install, uninstall, upgrade, etc.)
  - Add provider detection for build tools, download utilities, and script interpreters
  - Integrate with existing service management and file/directory handling
  - Add support for --provider flag and provider selection logic
  - _Requirements: 6.1, 6.2, 6.3, 6.5, 6.6_

- [x] 4.2 Add logging, debugging, and dry-run support
  - Implement detailed logging for all alternative provider actions
  - Add debug output for template resolution and command execution
  - Implement dry-run mode for showing planned actions without execution
  - Add progress indicators for long-running operations (builds, downloads)
  - _Requirements: 6.7, 6.8_

### Phase 5: Sample Configurations and Testing

- [x] 5. Create comprehensive sample saidata configurations
  - Update existing nginx sample with complete source build configuration
  - Create terraform sample demonstrating binary download configuration
  - Create docker sample demonstrating script installation configuration
  - Add OS-specific override examples for all three installation methods
  - _Requirements: 1.8, 2.8, 3.8, 5.7_

- [x] 5.1 Implement comprehensive test suite
  - Create unit tests for all new type definitions and methods
  - Add template function tests with various input combinations and edge cases
  - Create integration tests for end-to-end provider execution
  - Add cross-platform compatibility tests for all three providers
  - _Requirements: 4.4, 4.8, 6.4_

- [x] 5.2 Add validation and error handling tests
  - Test template resolution failure scenarios and graceful degradation
  - Add tests for invalid saidata configurations and schema validation
  - Test rollback procedures for failed installations
  - Add security tests for script execution and binary verification
  - _Requirements: 1.4, 2.7, 3.7, 4.4_

### Phase 6: Documentation and Examples

- [x] 6. Update documentation and examples
  - Update sai_source_functions.md with complete field reference
  - Create sai_binary_functions.md and sai_script_functions.md documentation
  - Add provider development guide for alternative installation methods
  - Update main SAI documentation with alternative provider examples
  - _Requirements: 4.6, 4.8_

- [x] 6.1 Create usage examples and tutorials
  - Write tutorial for building software from source with SAI
  - Create guide for downloading and installing binaries
  - Add examples for using script-based installations safely
  - Document best practices for saidata configuration with alternative providers
  - _Requirements: 6.1, 6.2, 6.3_

## Implementation Notes

### Build System Support
The source provider must support multiple build systems with appropriate defaults:
- **autotools**: `./configure && make && make install`
- **cmake**: `cmake . && cmake --build . && cmake --install .`
- **make**: Direct make-based builds
- **meson**: `meson setup build && meson compile -C build && meson install -C build`
- **ninja**: `ninja && ninja install`
- **custom**: User-defined build commands

### Security Considerations
- All downloads must support checksum verification (SHA256 recommended)
- Script execution must be opt-in with clear security warnings
- HTTPS enforcement for all download URLs
- Sandbox considerations for script execution environments

### Cross-Platform Compatibility
- OS detection for platform-specific configurations
- Architecture detection for binary downloads
- Path handling differences between Unix and Windows
- Permission setting compatibility across platforms

### Performance Optimization
- Parallel downloads where possible
- Build caching for source builds
- Incremental updates for binary installations
- Progress reporting for long-running operations

## Success Criteria

1. **Functional**: All three providers (source, binary, script) work end-to-end
2. **Compatible**: Seamless integration with existing SAI features and providers
3. **Secure**: Proper validation, checksums, and safe execution practices
4. **Documented**: Complete documentation and examples for all new features
5. **Tested**: Comprehensive test coverage including edge cases and error scenarios
6. **Cross-platform**: Works correctly on Linux, macOS, and Windows