# BOBA - CLI Tool Installer

> **Built with ❤️ using [Kiro IDE](https://kiro.ai) - A vibe code project**

BOBA is an interactive command-line tool installer that provides a comprehensive solution for setting up development environments across different platforms (WSL, Linux, and macOS). It uses a GitHub-driven configuration approach where all tool definitions and setup instructions are stored in your private repository, with local override capabilities for selective installation.

## ✨ Features

### 🎯 Core Functionality
- **Interactive Menu System**: Navigate through options using arrow keys with a beautiful terminal UI
- **GitHub Integration**: Authenticate and fetch configurations from your private repository
- **Cross-Platform Support**: Works seamlessly on Linux, macOS, and WSL
- **Smart Installation**: Detects already installed tools and avoids duplicates
- **Local Overrides**: Customize which tools to install while keeping your GitHub config intact
- **Environment Setup**: Automated shell configuration management (zsh)
- **Installation History**: Track installed tools with versions and timestamps
- **Update Management**: Update all previously installed tools with a single command

### 🛠️ Tool Management
- **Install Everything**: One-click installation of your entire development stack
- **Selective Installation**: Choose specific tools from your curated list
- **Installation Verification**: Automatic verification that tools are properly installed
- **Uninstall Support**: Clean removal of tools when needed
- **Status Tracking**: Visual indicators for installation status

### 🔧 Configuration Management
- **Tool Overrides**: Enable/disable specific tools locally
- **Environment Overrides**: Control which shell configurations to apply
- **Repository Settings**: Configure GitHub repository URL and authentication
- **Persistent Settings**: All configurations saved locally for future sessions

### 🌍 Environment Setup
- **Shell Configuration**: Automated zsh configuration from your repository
- **Multiple Environments**: Support for different shell setups (development, minimal, etc.)
- **Auto-Apply Settings**: Automatically apply environment configurations during "Install Everything"
- **Backup Management**: Automatic backup of existing configurations before changes

## 🚀 Quick Start

### Prerequisites
- Go 1.24+ installed on your system
- Git installed and configured
- zsh shell (for environment setup features)
- GitHub personal access token with repository access

### Installation

#### Option 1: Install to System (Recommended)
1. Download or clone this repository
2. Build the application:
   ```bash
   go build -o boba
   ```
3. Run BOBA and select "Install BOBA to System" from the main menu
4. This will install BOBA to `/usr/local/bin` and set up shell integration

#### Option 2: Manual Installation
1. Build the application:
   ```bash
   go build -o boba
   ```
2. Move to your PATH:
   ```bash
   sudo mv boba /usr/local/bin/
   ```
3. Make executable:
   ```bash
   sudo chmod +x /usr/local/bin/boba
   ```

### First Run
1. Run `boba` in your terminal
2. Navigate to "Installation Configuration" → "GitHub Repository Settings"
3. Enter your GitHub repository URL (e.g., `https://github.com/username/boba-config`)
4. Provide your GitHub personal access token
5. Return to main menu and start using BOBA!

## 📖 Usage

### Main Menu Options

```
┌─────────────────────────────────────┐
│              BOBA v1.0              │
│        CLI Tool Installer           │
├─────────────────────────────────────┤
│ → Install Everything                │
│   List of Available Tools           │
│   Setup Environment                 │
│   Installation Configuration        │
│   Install BOBA to System           │
│   Update Everything                 │
└─────────────────────────────────────┘
```

#### 🎯 Install Everything
Installs all tools from your GitHub repository configuration, respecting local overrides. Also applies auto-apply environment configurations.

#### 📋 List of Available Tools
Browse and selectively install tools from your repository. Shows installation status and allows individual tool management.

#### 🌍 Setup Environment
Configure your shell environment with custom configurations from your repository.

#### ⚙️ Installation Configuration
- **Tool Installation Overrides**: Enable/disable specific tools
- **Environment Overrides**: Control environment configurations
- **GitHub Repository Settings**: Configure repository URL and authentication

#### 🔄 Update Everything
Updates all previously installed tools to their latest versions.

#### 🔧 Install BOBA to System
Installs BOBA to your system PATH (`/usr/local/bin`) and sets up shell integration:
- Copies binary to system location with proper permissions
- Adds PATH configuration to `~/.zshrc`
- Creates helpful aliases (`boba-update`, `dev-setup`)
- Handles sudo requirements automatically
- Creates backups before making changes

### Navigation
- **Arrow Keys**: Navigate menu options
- **Enter**: Select menu item
- **'b' or Escape**: Go back to previous menu
- **'q'**: Quit application
- **Ctrl+C**: Force quit

## 🎬 Demo

Here's what the BOBA interface looks like in action:

```
┌─────────────────────────────────────────────────────────────┐
│                        BOBA v1.0                            │
│                   CLI Tool Installer                        │
│                                                             │
│  Built with ❤️ using Kiro IDE                              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  → 🚀 Install Everything                                    │
│    📋 List of Available Tools                               │
│    🌍 Setup Environment                                     │
│    ⚙️  Installation Configuration                           │
│    🔧 Install BOBA to System                               │
│    🔄 Update Everything                                     │
│                                                             │
│  Use ↑↓ to navigate, Enter to select, 'q' to quit         │
└─────────────────────────────────────────────────────────────┘
```

### Tool Installation Progress
```
┌─────────────────────────────────────────────────────────────┐
│                   Installing Tools...                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ✅ Node.js v18.17.0 - Installed successfully              │
│  ✅ Docker v24.0.6 - Installed successfully                │
│  🔄 AWS CLI - Installing...                                │
│  ⬜ Python 3.11 - Pending                                  │
│  ⬜ Git - Pending                                           │
│                                                             │
│  Progress: 2/5 tools completed                             │
└─────────────────────────────────────────────────────────────┘
```

### Available Tools List
```
┌─────────────────────────────────────────────────────────────┐
│                    Available Tools                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  → ✅ Node.js v18.17.0        JavaScript runtime           │
│    ⬜ Docker                  Container platform            │
│    ✅ AWS CLI v2.13.0         Amazon Web Services CLI      │
│    ⬜ Python 3.11             Programming language          │
│    ✅ Git v2.42.0             Version control system       │
│                                                             │
│  ✅ = Installed  ⬜ = Not Installed  ⚡ = Auto-install     │
│  Use ↑↓ to navigate, Enter to install, 'b' to go back     │
└─────────────────────────────────────────────────────────────┘
```

## 🗂️ Repository Configuration

BOBA works with a GitHub repository that contains your tool and environment configurations. Here's the expected structure:

```
your-boba-config/
├── tools/
│   ├── nodejs/
│   │   ├── tool.yaml
│   │   ├── install.sh
│   │   └── uninstall.sh
│   ├── docker/
│   │   ├── tool.yaml
│   │   ├── install.sh
│   │   └── uninstall.sh
│   └── aws-cli/
│       ├── tool.yaml
│       ├── install.sh
│       └── uninstall.sh
└── environments/
    ├── zsh-dev/
    │   ├── environment.yaml
    │   └── .zshrc
    └── zsh-minimal/
        ├── environment.yaml
        └── .zshrc
```

### Tool Configuration (tool.yaml)
```yaml
name: "Node.js"
description: "JavaScript runtime environment"
category: "development"
auto_install: true
check_command: "node --version"
```

### Environment Configuration (environment.yaml)
```yaml
name: "Development ZSH"
description: "Full development environment with plugins"
shell: "zsh"
auto_apply: true
config_files:
  - ".zshrc"
```

For detailed configuration guide, see [BOBA_CONFIG_GUIDE.md](BOBA_CONFIG_GUIDE.md).

## 🔧 Configuration Files

BOBA stores its configuration in `~/.boba/`:

```
~/.boba/
├── config.json          # Main configuration
├── credentials.json     # GitHub authentication (secure)
└── cache/              # Repository cache
    └── tools.json
```

### config.json
```json
{
  "repository_url": "https://github.com/username/boba-config",
  "tool_overrides": {
    "docker": false,
    "nodejs": true
  },
  "environment_overrides": {
    "zsh-dev": true,
    "zsh-minimal": false
  },
  "installed_tools": {
    "nodejs": {
      "name": "nodejs",
      "version": "v18.17.0",
      "install_date": "2024-10-25T10:30:00Z",
      "install_method": "auto"
    }
  },
  "last_sync": "2024-10-25T10:30:00Z"
}
```

## 🛠️ Development

### Building from Source
```bash
# Clone the repository
git clone <repository-url>
cd boba

# Install dependencies
go mod tidy

# Build the application
go build -o boba

# Run tests
go test ./...
```

### Project Structure
```
boba/
├── main.go                 # Application entry point
├── internal/
│   ├── ui/                # Bubble Tea UI components
│   ├── github/            # GitHub API integration
│   ├── config/            # Configuration management
│   ├── installer/         # Installation engine
│   └── parser/            # Repository parsing
├── cmd/                   # Command-line tools
└── .kiro/                 # Kiro IDE specifications
    └── specs/
        └── cli-tool-installer/
```

### Dependencies
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)**: Terminal UI framework
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)**: Terminal styling
- **[go-github](https://github.com/google/go-github)**: GitHub API client
- **[oauth2](https://golang.org/x/oauth2)**: OAuth2 authentication
- **[yaml.v3](https://gopkg.in/yaml.v3)**: YAML parsing

## 🔍 Troubleshooting

### Common Issues

#### Authentication Problems
**Problem**: "Failed to authenticate with GitHub"
**Solution**: 
1. Verify your personal access token has repository access
2. Check if the repository URL is correct
3. Ensure the repository exists and is accessible

#### Installation Failures
**Problem**: Tool installation fails
**Solution**:
1. Check if you have necessary permissions (sudo for system packages)
2. Verify your platform is supported by the tool's install script
3. Check network connectivity for downloads

#### Repository Access Issues
**Problem**: "Cannot access repository"
**Solution**:
1. Verify the repository URL format: `https://github.com/username/repo-name`
2. Ensure your GitHub token has access to the repository
3. Check if the repository is private and token has appropriate permissions

#### Platform Detection Issues
**Problem**: "Unsupported platform"
**Solution**:
1. BOBA supports Linux, macOS, and WSL
2. Ensure your platform has the required package managers (apt, brew, etc.)
3. Check if the tool's install script supports your specific platform

#### System Installation Issues
**Problem**: "Permission denied" during system installation
**Solution**:
1. Ensure you have sudo privileges on your system
2. Check that `/usr/local/bin` is writable or exists
3. Try running with elevated privileges if on Windows

#### Environment Setup Problems
**Problem**: Shell configuration not applied
**Solution**:
1. Restart your terminal or run `source ~/.zshrc`
2. Verify zsh is your default shell: `echo $SHELL`
3. Check if `.zshrc` backup was created and restore if needed

#### Tool Update Issues
**Problem**: "Update Everything" not finding tools
**Solution**:
1. Ensure tools were installed through BOBA (check installation history)
2. Verify GitHub repository access for latest tool definitions
3. Check if local overrides are preventing updates

### Debug Mode
Run BOBA with debug information:
```bash
# Enable verbose logging
export BOBA_DEBUG=1
boba

# Check installation logs
cat ~/.boba/logs/installation.log
```

### Reset Configuration
If you encounter persistent issues, you can reset BOBA's configuration:
```bash
# Backup current config
cp -r ~/.boba ~/.boba.backup

# Remove configuration (will prompt for setup on next run)
rm -rf ~/.boba

# Or just reset overrides
rm ~/.boba/config.json
```

### Getting Help
1. Check the [Issues](https://github.com/your-repo/boba/issues) page
2. Review the configuration guide: [BOBA_CONFIG_GUIDE.md](BOBA_CONFIG_GUIDE.md)
3. Check integration notes: [INTEGRATION_NOTES.md](INTEGRATION_NOTES.md)

## 📋 FAQ

### Q: Can I use BOBA without a GitHub repository?
A: No, BOBA is designed to work with GitHub repositories for configuration management. This ensures your setup is version-controlled and shareable.

### Q: Does BOBA work on Windows?
A: BOBA is primarily designed for Unix-like systems (Linux, macOS, WSL). Native Windows support is not currently available.

### Q: Can I contribute my own tools?
A: Yes! Create install/uninstall scripts in your repository following the configuration guide. BOBA will automatically detect and use them.

### Q: How do I backup my configuration?
A: Your main configuration is in your GitHub repository. Local overrides are stored in `~/.boba/config.json` - back this up if you have custom local settings.

### Q: Can I use BOBA in CI/CD pipelines?
A: While BOBA is designed for interactive use, you could potentially script it for automated environment setup. However, consider using the underlying scripts directly for CI/CD.

### Q: How do I update BOBA itself?
A: Currently, you need to rebuild from source. Future versions will include self-update functionality.

### Q: Can I use BOBA with public repositories?
A: Yes! While BOBA was designed for private repositories, it works perfectly with public ones. Just ensure your repository follows the expected structure.

### Q: What happens if I lose my GitHub token?
A: Simply navigate to "Installation Configuration" → "GitHub Repository Settings" and enter a new token. Your local overrides and installed tool history will be preserved.

### Q: Can I have different configurations for different machines?
A: Yes! Use local overrides to customize which tools are installed on each machine while keeping your main configuration in GitHub.

### Q: Does BOBA support custom package managers?
A: BOBA detects common package managers (apt, yum, brew, etc.) automatically. For custom package managers, you'll need to handle them in your tool's install scripts.

### Q: How do I share my BOBA configuration with my team?
A: Share your GitHub repository URL! Team members can use the same repository with their own local overrides as needed.

### Q: Can I run BOBA in headless mode?
A: BOBA is designed for interactive use. For automated setups, consider using the underlying scripts directly or creating wrapper scripts.

### Q: What if a tool installation fails?
A: BOBA provides detailed error messages and logs. You can retry individual tools or check the installation logs in `~/.boba/logs/` for debugging.

## 🤝 Contributing

We welcome contributions! This project was built using [Kiro IDE](https://kiro.ai) as a vibe code project, showcasing rapid development with AI assistance.

### Development Workflow
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

### Code Style
- Follow Go conventions and best practices
- Use `gofmt` for formatting
- Add comments for public functions
- Include tests for new features

## 🚀 Roadmap

### Upcoming Features
- **Self-Update Functionality**: Update BOBA itself from within the application
- **Plugin System**: Support for custom tool installers and extensions
- **Configuration Templates**: Pre-built configurations for common development stacks
- **Parallel Installations**: Install multiple tools simultaneously for faster setup
- **Dependency Management**: Automatic handling of tool dependencies
- **Custom Environments**: Support for more shell types (fish, bash, etc.)
- **Installation Profiles**: Different installation profiles for different use cases
- **Rollback System**: Easy rollback of installations and configurations

### Long-term Vision
- **Team Management**: Shared team configurations with role-based access
- **Cloud Sync**: Synchronize configurations across multiple machines
- **Integration APIs**: REST API for integration with other tools
- **GUI Version**: Desktop application for users who prefer graphical interfaces
- **Package Registry**: Community-driven registry of tool configurations

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🎯 Project Story

BOBA was created as a **vibe code project** using [Kiro IDE](https://kiro.ai), showcasing the power of AI-assisted development. The entire application was built through iterative collaboration between human creativity and AI capabilities, demonstrating how modern development tools can accelerate the creation of complex, feature-rich applications.

### Development Highlights
- **Spec-Driven Development**: Used Kiro's specification system to plan and implement features systematically
- **AI-Assisted Coding**: Leveraged AI for rapid prototyping and implementation
- **Interactive Design**: Built with user experience as the primary focus
- **Cross-Platform Thinking**: Designed from the ground up to work across different environments

This project serves as an example of what's possible when developers combine their domain expertise with AI-powered development tools.

## 🙏 Acknowledgments

- **[Kiro IDE](https://kiro.ai)** - The AI-powered development environment that made this project possible
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - For the beautiful and responsive terminal UI framework
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - For elegant terminal styling and layout
- **[go-github](https://github.com/google/go-github)** - For seamless GitHub API integration
- **The Go Community** - For creating an excellent ecosystem for CLI applications
- **Open Source Contributors** - For the amazing tools and libraries that make projects like this possible

## 🌟 Why "BOBA"?

BOBA stands for **"Build Once, Bootstrap Anywhere"** - reflecting the core philosophy of setting up your development environment once in a repository and being able to bootstrap it anywhere you need it. Just like the delightful drink, BOBA brings together different ingredients (tools) to create something greater than the sum of its parts.

---

**Made with ❤️ and AI assistance through [Kiro IDE](https://kiro.ai)**  
*A vibe code project demonstrating the future of collaborative development*