# Contributing to SAI CLI

Thank you for your interest in contributing to SAI CLI! This guide will help you get started with contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Ways to Contribute](#ways-to-contribute)
- [Development Setup](#development-setup)
- [Contributing Code](#contributing-code)
- [Contributing Providers](#contributing-providers)
- [Contributing Documentation](#contributing-documentation)
- [Submitting Changes](#submitting-changes)
- [Review Process](#review-process)
- [Community](#community)

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to [sai@example42.com](mailto:sai@example42.com).

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Make
- Basic understanding of YAML and command-line tools

### Quick Start

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/example42/sai.git
   cd sai
   ```
3. **Set up the development environment**:
   ```bash
   make deps
   make build
   make test
   ```
4. **Create a branch** for your changes:
   ```bash
   git checkout -b feature/my-new-feature
   ```

## Ways to Contribute

### 🐛 Report Bugs

Found a bug? Please create an issue with:
- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- System information (`sai stats --system --verbose`)
- SAI version (`sai version --verbose`)

### 💡 Suggest Features

Have an idea? Start a discussion:
- Describe the use case
- Explain the expected behavior
- Consider implementation approaches
- Discuss potential alternatives

### 📝 Improve Documentation

Documentation improvements are always welcome:
- Fix typos and grammar
- Add examples and use cases
- Improve clarity and organization
- Translate to other languages

### 🔧 Add Providers

Providers are the easiest way to contribute:
- Add support for new package managers
- Create specialized tool providers
- Improve existing provider configurations
- Add OS-specific overrides

### 💻 Contribute Code

Code contributions include:
- Bug fixes
- New features
- Performance improvements
- Test coverage improvements
- Refactoring and cleanup

## Development Setup

### Environment Setup

1. **Install Go 1.21+**:
   ```bash
   # Check version
   go version
   ```

2. **Install development tools**:
   ```bash
   make deps
   ```

3. **Verify setup**:
   ```bash
   make verify-env
   make test
   make lint
   ```

### Project Structure

```
├── cmd/sai/                   # Main application entry point
├── internal/                  # Private application code
│   ├── cli/                   # CLI command implementations
│   ├── action/                # Action management
│   ├── provider/              # Provider loading and management
│   ├── saidata/               # Software data management
│   ├── template/              # Template engine
│   └── ...                    # Other internal packages
├── docs/                      # Documentation
├── providers/                 # Provider YAML files
├── schemas/                   # JSON Schema files
├── scripts/                   # Build and installation scripts
└── .github/workflows/         # CI/CD workflows
```

### Development Workflow

1. **Make changes** in your feature branch
2. **Add tests** for new functionality
3. **Run tests** locally:
   ```bash
   make test
   make lint
   make build
   ```
4. **Test manually**:
   ```bash
   ./build/sai install nginx --dry-run --verbose
   ```
5. **Commit changes** with clear messages
6. **Push to your fork** and create a pull request

## Contributing Code

### Coding Standards

- **Go Style**: Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- **Formatting**: Use `gofmt` and `goimports`
- **Linting**: Pass `golangci-lint` checks
- **Testing**: Maintain or improve test coverage
- **Documentation**: Include godoc comments for public APIs

### Writing Tests

```go
// Example unit test
func TestProviderManager_LoadProviders(t *testing.T) {
    manager := NewProviderManager()
    
    err := manager.LoadProviders("testdata/providers")
    assert.NoError(t, err)
    
    providers := manager.GetAvailableProviders()
    assert.NotEmpty(t, providers)
}

// Example integration test
func TestCLI_InstallCommand(t *testing.T) {
    cmd := exec.Command("./build/sai", "install", "nginx", "--dry-run")
    output, err := cmd.CombinedOutput()
    
    assert.NoError(t, err)
    assert.Contains(t, string(output), "apt install")
}
```

### Adding New Commands

1. **Create command file** in `internal/cli/`:
   ```go
   // internal/cli/mycommand.go
   func NewMyCommand() *cobra.Command {
       cmd := &cobra.Command{
           Use:   "mycommand",
           Short: "Description of my command",
           RunE:  runMyCommand,
       }
       return cmd
   }
   ```

2. **Add to root command** in `internal/cli/root.go`:
   ```go
   rootCmd.AddCommand(NewMyCommand())
   ```

3. **Write tests** in `internal/cli/mycommand_test.go`

4. **Update documentation**

### Error Handling

Use structured errors with context:

```go
import "github.com/pkg/errors"

func (m *Manager) LoadProvider(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return errors.Wrapf(err, "failed to read provider file: %s", path)
    }
    
    var provider ProviderData
    if err := yaml.Unmarshal(data, &provider); err != nil {
        return errors.Wrapf(err, "failed to parse provider YAML: %s", path)
    }
    
    return nil
}
```

## Contributing Providers

Providers are the easiest way to contribute! They're defined in YAML files and don't require Go knowledge.

### Creating a New Provider

1. **Choose provider type**:
   - Package manager (apt, dnf, etc.)
   - Container platform (docker, podman)
   - Language package manager (npm, pip)
   - Specialized tool (security, monitoring)

2. **Create provider file**:
   ```yaml
   # providers/my-provider.yaml
   version: "1.0"
   provider:
     name: "my-provider"
     display_name: "My Package Manager"
     type: "package_manager"
     platforms: ["linux"]
     executable: "mypkg"
     priority: 50
   
   actions:
     install:
       description: "Install package"
       template: "mypkg install {{sai_package}}"
       requires_root: true
       timeout: 300
       validation:
         command: "mypkg list | grep -q '^{{sai_package}}$'"
       rollback: "mypkg remove {{sai_package}}"
   ```

3. **Test the provider**:
   ```bash
   # Validate syntax
   sai stats --provider my-provider
   
   # Test with dry-run
   sai install nginx --provider my-provider --dry-run
   ```

4. **Add documentation** and examples

### Provider Best Practices

- **Use descriptive names** and clear descriptions
- **Support multiple platforms** when possible
- **Include validation commands** for safety
- **Provide rollback procedures** for destructive actions
- **Test on target platforms** thoroughly
- **Follow existing provider patterns**

### Specialized Providers

For debugging, security, monitoring, and other specialized tools:

```yaml
# providers/specialized/security-scanner.yaml
version: "1.0"
provider:
  name: "security-scanner"
  display_name: "Security Scanner"
  type: "security"
  platforms: ["linux", "darwin"]
  executable: "scanner"

actions:
  check:
    description: "Run security scan"
    template: "scanner scan {{sai_file 'binary'}} --format json"
    timeout: 300
  
  info:
    description: "Show security information"
    template: "scanner info {{.Software}}"
```

## Contributing Documentation

### Types of Documentation

- **User guides**: Help users accomplish tasks
- **API documentation**: Document code interfaces
- **Examples**: Show real-world usage
- **Troubleshooting**: Help solve common problems

### Documentation Standards

- **Clear and concise**: Use simple language
- **Practical examples**: Include working code samples
- **Up-to-date**: Keep in sync with code changes
- **Well-organized**: Use consistent structure
- **Accessible**: Consider different skill levels

### Writing Examples

```markdown
### Installing Web Servers

Install and start Nginx:

```bash
# Install Nginx
sai install nginx

# Start the service
sai start nginx

# Enable at boot
sai enable nginx

# Check status
sai status nginx
```

This will:
1. Detect the best package manager for your system
2. Install the nginx package
3. Start the nginx service
4. Configure it to start automatically at boot
5. Show the current service status
```

## Submitting Changes

### Pull Request Process

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make your changes** with clear, focused commits:
   ```bash
   git add .
   git commit -m "feat: add support for new package manager"
   ```

3. **Push to your fork**:
   ```bash
   git push origin feature/my-feature
   ```

4. **Create a pull request** on GitHub with:
   - Clear title and description
   - Reference to related issues
   - Screenshots/examples if applicable
   - Checklist of changes made

### Commit Message Format

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test additions/changes
- `chore`: Maintenance tasks

Examples:
```
feat(provider): add support for pacman package manager
fix(cli): resolve template resolution error for missing saidata
docs(examples): add MongoDB configuration examples
```

### Pull Request Checklist

- [ ] Code follows project style guidelines
- [ ] Self-review of code completed
- [ ] Tests added for new functionality
- [ ] All tests pass locally
- [ ] Documentation updated
- [ ] Commit messages follow convention
- [ ] No merge conflicts with main branch

## Review Process

### What to Expect

1. **Automated checks**: CI/CD runs tests and linting
2. **Maintainer review**: Code review by project maintainers
3. **Feedback incorporation**: Address review comments
4. **Final approval**: Merge when ready

### Review Criteria

- **Functionality**: Does it work as intended?
- **Code quality**: Is it well-written and maintainable?
- **Testing**: Are there adequate tests?
- **Documentation**: Is it properly documented?
- **Compatibility**: Does it work across platforms?
- **Performance**: Does it impact performance?

### Addressing Feedback

- **Be responsive**: Address feedback promptly
- **Ask questions**: Clarify unclear feedback
- **Make changes**: Update code based on suggestions
- **Test thoroughly**: Ensure changes work correctly
- **Update documentation**: Keep docs in sync

## Community

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and general discussion
- **Pull Requests**: Code review and collaboration
- **Email**: [sai@example42.com](mailto:sai@example42.com) for private matters

### Getting Help

- **Documentation**: Check existing docs first
- **Search issues**: Look for similar problems
- **Ask questions**: Use GitHub Discussions
- **Join discussions**: Participate in community conversations

### Recognition

Contributors are recognized through:
- **Contributors list**: Listed in README and releases
- **Release notes**: Contributions mentioned in changelogs
- **Special thanks**: Recognition for significant contributions

## Development Tips

### Useful Commands

```bash
# Development workflow
make build-dev          # Build with race detection
make test               # Run all tests
make test-coverage      # Run tests with coverage
make lint               # Run linter
make fmt                # Format code

# Testing specific components
go test ./internal/provider/...
go test -run TestProviderManager

# Manual testing
./build/sai install nginx --dry-run --verbose
./build/sai stats --providers
```

### Debugging

```bash
# Enable debug logging
export SAI_LOG_LEVEL=debug
./build/sai install nginx --verbose

# Use delve debugger
dlv debug ./cmd/sai -- install nginx --dry-run
```

### Performance Testing

```bash
# Benchmark tests
go test -bench=. ./internal/...

# Memory profiling
go test -memprofile=mem.prof ./internal/provider/
go tool pprof mem.prof
```

## Thank You!

Your contributions make SAI CLI better for everyone. Whether you're fixing a typo, adding a provider, or implementing a major feature, every contribution is valuable and appreciated.

Happy contributing! 🚀

---

For questions about contributing, please:
- 📖 Read the [documentation](README.md)
- 💬 Start a [discussion](https://github.com/example42/sai/discussions)
- 📧 Email us at [sai@example42.com](mailto:sai@example42.com)