# BOBA Configuration Repository Guide

This guide explains how to create a `boba-config` repository that works with the BOBA CLI Tool Installer.

## Repository Structure

Your `boba-config` repository should follow this structure:

```
boba-config/
├── tools/
│   ├── nodejs/
│   │   ├── tool.yaml
│   │   ├── install.sh
│   │   └── uninstall.sh
│   ├── docker/
│   │   ├── tool.yaml
│   │   ├── install.sh
│   │   └── uninstall.sh
│   ├── aws-cli/
│   │   ├── tool.yaml
│   │   ├── install.sh
│   │   └── uninstall.sh
│   └── python/
│       ├── tool.yaml
│       ├── install.sh
│       └── uninstall.sh
├── environments/
│   ├── zsh-dev/
│   │   ├── environment.yaml
│   │   ├── .zshrc
│   │   ├── setup.sh
│   │   └── restore.sh
│   ├── bash-minimal/
│   │   ├── environment.yaml
│   │   ├── .bashrc
│   │   ├── setup.sh
│   │   └── restore.sh
│   └── fish-config/
│       ├── environment.yaml
│       ├── .fishrc
│       ├── setup.sh
│       └── restore.sh
└── README.md
```

## Tool Configuration

Each tool folder must contain three files:

### 1. `tool.yaml` (or `tool.json`)

This file contains metadata about the tool:

```yaml
name: "Node.js"
description: "JavaScript runtime environment"
version: "20.x LTS"
auto_install: true
dependencies: []
homepage: "https://nodejs.org"
```

**Field Descriptions:**
- `name`: Display name of the tool
- `description`: Brief description of what the tool does
- `version`: Version or version range to install (optional)
- `auto_install`: If `true`, installs automatically with "Install Everything". If `false`, user must install manually from tools list
- `dependencies`: List of other tools this tool depends on (optional)
- `homepage`: Official website URL (optional)

**Alternative JSON format:**
```json
{
  "name": "Node.js",
  "description": "JavaScript runtime environment",
  "version": "20.x LTS",
  "auto_install": true,
  "dependencies": [],
  "homepage": "https://nodejs.org"
}
```

### 2. `install.sh`

This script handles the installation of the tool:

```bash
#!/bin/bash

# Node.js Installation Script
set -e

echo "Installing Node.js..."

# Detect platform
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux installation
    curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
    sudo apt-get install -y nodejs
elif [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS installation
    if command -v brew &> /dev/null; then
        brew install node
    else
        echo "Homebrew not found. Please install Homebrew first."
        exit 1
    fi
elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
    # Windows installation
    echo "Please download Node.js from https://nodejs.org"
    exit 1
else
    echo "Unsupported platform: $OSTYPE"
    exit 1
fi

# Verify installation
if command -v node &> /dev/null; then
    echo "Node.js installed successfully!"
    node --version
    npm --version
else
    echo "Node.js installation failed!"
    exit 1
fi
```

### 3. `uninstall.sh`

This script handles the removal of the tool:

```bash
#!/bin/bash

# Node.js Uninstallation Script
set -e

echo "Uninstalling Node.js..."

# Detect platform
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux uninstallation
    sudo apt-get remove -y nodejs npm
    sudo apt-get autoremove -y
elif [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS uninstallation
    if command -v brew &> /dev/null; then
        brew uninstall node
    else
        echo "Homebrew not found. Manual removal may be required."
    fi
elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
    # Windows uninstallation
    echo "Please uninstall Node.js from Windows Settings > Apps"
    exit 1
else
    echo "Unsupported platform: $OSTYPE"
    exit 1
fi

echo "Node.js uninstalled successfully!"
```

## Example Tools

Here are some example tool configurations:

### Docker

**`tools/docker/tool.yaml`:**
```yaml
name: "Docker"
description: "Container platform for building and running applications"
version: "latest"
auto_install: true
dependencies: []
homepage: "https://docker.com"
```

**`tools/docker/install.sh`:**
```bash
#!/bin/bash
set -e

echo "Installing Docker..."

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Install Docker on Linux
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker $USER
    rm get-docker.sh
elif [[ "$OSTYPE" == "darwin"* ]]; then
    # Install Docker Desktop on macOS
    if command -v brew &> /dev/null; then
        brew install --cask docker
    else
        echo "Please download Docker Desktop from https://docker.com"
        exit 1
    fi
fi

echo "Docker installed successfully!"
```

### AWS CLI

**`tools/aws-cli/tool.yaml`:**
```yaml
name: "AWS CLI"
description: "Command line interface for Amazon Web Services"
version: "2.x"
auto_install: false
dependencies: []
homepage: "https://aws.amazon.com/cli/"
```

**`tools/aws-cli/install.sh`:**
```bash
#!/bin/bash
set -e

echo "Installing AWS CLI v2..."

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux installation
    curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
    unzip awscliv2.zip
    sudo ./aws/install
    rm -rf awscliv2.zip aws/
elif [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS installation
    curl "https://awscli.amazonaws.com/AWSCLIV2.pkg" -o "AWSCLIV2.pkg"
    sudo installer -pkg AWSCLIV2.pkg -target /
    rm AWSCLIV2.pkg
fi

# Verify installation
if command -v aws &> /dev/null; then
    echo "AWS CLI installed successfully!"
    aws --version
else
    echo "AWS CLI installation failed!"
    exit 1
fi
```

## Best Practices

### Script Guidelines

1. **Always use `set -e`** to exit on errors
2. **Detect the platform** before installation
3. **Verify installation** at the end
4. **Provide clear error messages**
5. **Clean up temporary files**
6. **Make scripts executable**: `chmod +x install.sh uninstall.sh`

### Platform Detection

Use these patterns for cross-platform support:

```bash
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux
elif [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
    # Windows (Git Bash/WSL)
else
    echo "Unsupported platform: $OSTYPE"
    exit 1
fi
```

### Package Manager Detection

```bash
# Check for package managers
if command -v apt-get &> /dev/null; then
    # Debian/Ubuntu
    sudo apt-get update
    sudo apt-get install -y package-name
elif command -v yum &> /dev/null; then
    # RHEL/CentOS
    sudo yum install -y package-name
elif command -v brew &> /dev/null; then
    # macOS with Homebrew
    brew install package-name
fi
```

## Repository Setup

1. **Create a new GitHub repository** named `boba-config` (recommended default name)
2. **Make it private** if you want to keep your configurations private
3. **Add the tools** you want to manage following the structure above
4. **Test your scripts** locally before committing

**Note:** BOBA automatically looks for a repository named `boba-config` by default. If you want to use a different name, you can change it in BOBA's "Installation Configuration" → "Repository Configuration" menu.

## Auto Install Behavior

The `auto_install` field controls whether a tool is installed automatically:

- **`auto_install: true`** - Tool will be installed when user selects "Install Everything"
- **`auto_install: false`** - Tool will only be available through "List of Available Tools" for manual installation

**Visual Indicators in BOBA:**
- ⚡ Lightning bolt icon = Auto-install tools (`auto_install: true`)
- 🔧 Wrench icon = Manual-install tools (`auto_install: false`)

**Example use cases for `auto_install: false`:**
- Optional tools that not everyone needs (AWS CLI, specialized databases)
- Tools that require additional configuration after installation
- Tools that might conflict with existing installations
- Large tools that users might want to install selectively

**Examples:**

```yaml
# Essential development tool - auto install
name: "Git"
description: "Version control system"
auto_install: true

# Optional cloud tool - manual install only
name: "AWS CLI"
description: "Amazon Web Services command line interface"
auto_install: false

# Core runtime - auto install
name: "Node.js"
description: "JavaScript runtime environment"
auto_install: true

# Specialized database - manual install only
name: "PostgreSQL"
description: "Advanced open source database"
auto_install: false
```

## Environment Configuration

The `environments/` directory allows you to manage shell configurations and development environments. Each environment folder contains configuration files and setup scripts for different shell environments.

### Environment Structure

Each environment folder must contain:

#### 1. `environment.yaml` (or `environment.json`)

This file contains metadata about the environment:

```yaml
name: "ZSH Development Environment"
description: "Optimized zsh configuration for development with plugins and themes"
shell: "zsh"
auto_apply: false
dependencies: ["git", "curl", "zsh"]
```

**Field Descriptions:**
- `name`: Display name of the environment configuration
- `description`: Brief description of what this environment provides
- `shell`: Target shell (zsh, bash, fish, etc.) - affects the icon displayed in BOBA
- `auto_apply`: If `true`, applies automatically with "Install Everything". If `false`, user must apply manually
- `dependencies`: List of tools that should be installed before applying this environment (optional)

#### 2. Shell Configuration Files

Include the actual configuration files for your shell:

- `.zshrc` - For zsh configurations
- `.bashrc` - For bash configurations  
- `.fishrc` - For fish shell configurations
- `.profile` - For general shell configurations
- `.bash_profile` - For bash login configurations

#### 3. `setup.sh`

Script to apply the environment configuration:

```bash
#!/bin/bash
set -e

echo "Setting up ZSH Development Environment..."

# Backup existing configuration
if [ -f "$HOME/.zshrc" ]; then
    cp "$HOME/.zshrc" "$HOME/.zshrc.backup.$(date +%Y%m%d_%H%M%S)"
    echo "Backed up existing .zshrc"
fi

# Copy new configuration
cp .zshrc "$HOME/.zshrc"
echo "Applied new .zshrc configuration"

# Install Oh My Zsh if not present
if [ ! -d "$HOME/.oh-my-zsh" ]; then
    sh -c "$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended
    echo "Installed Oh My Zsh"
fi

echo "Environment setup complete! Please restart your shell or run 'source ~/.zshrc'"
```

#### 4. `restore.sh`

Script to restore the previous configuration:

```bash
#!/bin/bash
set -e

echo "Restoring previous ZSH configuration..."

# Find the most recent backup
BACKUP_FILE=$(ls -t "$HOME"/.zshrc.backup.* 2>/dev/null | head -n1)

if [ -n "$BACKUP_FILE" ]; then
    cp "$BACKUP_FILE" "$HOME/.zshrc"
    echo "Restored configuration from $BACKUP_FILE"
else
    echo "No backup found. Creating minimal .zshrc"
    echo "# Minimal zsh configuration" > "$HOME/.zshrc"
fi

echo "Configuration restored! Please restart your shell."
```

### Environment Types and Icons

BOBA displays different icons based on the `shell` field:

- 🦓 **zsh** - Z shell configurations
- 🐚 **bash** - Bash shell configurations  
- 🐟 **fish** - Fish shell configurations
- 💻 **other** - Other or generic shell configurations

### Auto Apply Behavior

The `auto_apply` field controls when an environment is applied:

- **`auto_apply: true`** - Environment will be applied when user selects "Install Everything"
- **`auto_apply: false`** - Environment will only be available through "Setup Environment" for manual application

**Visual Indicators in BOBA:**
- ⚡ Lightning bolt icon = Auto-apply environments (`auto_apply: true`)
- 🔧 Wrench icon = Manual-apply environments (`auto_apply: false`)

### Example Environment Configurations

#### Minimal Bash Environment
```yaml
name: "Minimal Bash"
description: "Clean bash configuration with essential aliases"
shell: "bash"
auto_apply: true
dependencies: []
```

#### Advanced ZSH Development Environment
```yaml
name: "ZSH Dev Pro"
description: "Full-featured zsh with Oh My Zsh, plugins, and development tools"
shell: "zsh"
auto_apply: false
dependencies: ["git", "curl", "nodejs"]
```

#### Fish Shell Configuration
```yaml
name: "Fish Shell Setup"
description: "Modern fish shell with custom functions and themes"
shell: "fish"
auto_apply: false
dependencies: ["fish"]
```

### Best Practices for Environments

1. **Always backup existing configurations** in your setup scripts
2. **Test on multiple platforms** (Linux, macOS, WSL)
3. **Include restoration scripts** for easy rollback
4. **Document dependencies** clearly in the environment.yaml
5. **Use descriptive names** that indicate the purpose
6. **Keep configurations modular** - separate concerns into different environments
7. **Test setup and restore scripts** thoroughly before committing

### Platform-Specific Considerations

```bash
# Detect platform in your setup scripts
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS specific setup
    echo "Setting up for macOS..."
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux specific setup
    echo "Setting up for Linux..."
elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
    # Windows/WSL specific setup
    echo "Setting up for Windows..."
fi
```

## Testing Your Configuration

1. **Test scripts locally** on your target platforms
2. **Verify tool.yaml syntax** using a YAML validator
3. **Check file permissions** (scripts should be executable)
4. **Test with BOBA** by setting your repository URL in the configuration

## Example Repository

You can find a complete example repository at: `https://github.com/yourusername/boba-config`

## Troubleshooting

### Common Issues

1. **Scripts not executable**: Run `chmod +x install.sh uninstall.sh`
2. **YAML syntax errors**: Use a YAML validator
3. **Platform detection fails**: Check `$OSTYPE` variable
4. **Permission denied**: Ensure scripts have proper permissions
5. **Network issues**: Add retry logic for downloads

### Debug Mode

Add debug output to your scripts:

```bash
#!/bin/bash
set -e

# Enable debug mode
if [[ "${DEBUG}" == "1" ]]; then
    set -x
fi

echo "Installing tool..."
# ... rest of script
```

Run with: `DEBUG=1 ./install.sh`

---

This guide should help you create a comprehensive `boba-config` repository that works seamlessly with the BOBA CLI Tool Installer!