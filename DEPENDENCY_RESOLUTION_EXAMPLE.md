# Dependency Resolution in BOBA

BOBA now supports dependency resolution for both tools and environments. This ensures that dependencies are installed in the correct order.

## How It Works

### Tool Dependencies

When you define a tool with dependencies in your `tool.yaml`:

```yaml
name: cli-a
description: A CLI tool that depends on AWS CLI
dependencies:
  - aws-cli
auto_install: true
```

BOBA will:
1. Check if `aws-cli` is already installed
2. If not, install `aws-cli` first
3. Then install `cli-a`

### Environment Dependencies

Similarly, for environments in `environment.yaml`:

```yaml
name: dev-env
description: Development environment that extends base environment
dependencies:
  - base-env
auto_apply: true
```

BOBA will:
1. Apply `base-env` first
2. Then apply `dev-env`

## Installation Order

### Individual Tool Installation
When you install a single tool, BOBA will:
1. Resolve all dependencies for that tool
2. Install dependencies first (in dependency order)
3. Install the requested tool last

### "Install Everything" Mode
When you use "Install Everything", BOBA will:
1. **Phase 1: Tools**
   - Resolve dependencies for all selected tools
   - Install tools in dependency order
2. **Phase 2: Environments**
   - Resolve dependencies for all selected environments
   - Apply environments in dependency order

## Example Scenarios

### Scenario 1: Simple Chain
```
Tools:
- aws-cli (no dependencies)
- cli-a (depends on aws-cli)
- independent-tool (no dependencies)

Installation order: aws-cli → cli-a → independent-tool
```

### Scenario 2: Complex Dependencies
```
Tools:
- base-tool (no dependencies)
- middleware-tool (depends on base-tool)
- cli-a (depends on middleware-tool, aws-cli)
- aws-cli (no dependencies)

Installation order: base-tool → aws-cli → middleware-tool → cli-a
```

### Scenario 3: Environment Dependencies
```
Environments:
- base-env (no dependencies)
- dev-env (depends on base-env)
- prod-env (depends on base-env)

Application order: base-env → dev-env → prod-env
```

## Error Handling

BOBA will detect and report:
- **Circular dependencies**: When tool A depends on tool B, and tool B depends on tool A
- **Missing dependencies**: When a tool depends on another tool that doesn't exist
- **Installation failures**: If a dependency fails to install, the dependent tool installation is aborted

## Configuration

Dependencies are defined in the tool/environment configuration files:

**Tool configuration (`tools/my-tool/tool.yaml`):**
```yaml
name: my-tool
description: My awesome tool
dependencies:
  - dependency-tool-1
  - dependency-tool-2
auto_install: true
```

**Environment configuration (`environments/my-env/environment.yaml`):**
```yaml
name: my-env
description: My development environment
dependencies:
  - base-env
auto_apply: true
```

## Environment Application Process

When BOBA applies an environment, it:

1. **Downloads the setup script** from `environments/{env-name}/setup.sh`
2. **Executes the setup script** with environment variables:
   - `BOBA_ENV_NAME`: The environment name
   - `BOBA_ENV_SHELL`: Target shell (zsh, bash, fish, etc.)
   - `BOBA_PLATFORM`: Operating system (linux, darwin, windows)
   - `BOBA_PACKAGE_MANAGER`: Available package manager (apt, yum, brew, etc.)
3. **Verifies the application** (placeholder implementation)

### Environment Structure

Each environment should have this structure in your repository:

```
environments/
├── my-env/
│   ├── environment.yaml    # Environment configuration
│   ├── setup.sh           # Setup script (required)
│   ├── restore.sh         # Restore/cleanup script (optional)
│   ├── .zshrc             # Shell config files (optional)
│   ├── .bashrc            # Shell config files (optional)
│   └── .profile           # Shell config files (optional)
```

### Setup Script Example

```bash
#!/bin/bash
# environments/dev-env/setup.sh

echo "Setting up development environment..."

# Install development tools
if command -v apt >/dev/null 2>&1; then
    sudo apt update
    sudo apt install -y git curl wget
elif command -v brew >/dev/null 2>&1; then
    brew install git curl wget
fi

# Set up shell configuration
if [ "$BOBA_ENV_SHELL" = "zsh" ]; then
    echo "export DEV_MODE=true" >> ~/.zshrc
    echo "alias ll='ls -la'" >> ~/.zshrc
elif [ "$BOBA_ENV_SHELL" = "bash" ]; then
    echo "export DEV_MODE=true" >> ~/.bashrc
    echo "alias ll='ls -la'" >> ~/.bashrc
fi

echo "Development environment setup complete!"
```

This ensures that your tools and environments are always installed in the correct order, preventing installation failures due to missing dependencies.