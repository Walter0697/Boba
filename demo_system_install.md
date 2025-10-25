# BOBA System Installation Demo

This document demonstrates the system installation functionality implemented for task 18.

## Features Implemented

### 1. Main Menu Integration
- Added "🔧 Install BOBA to System" option to the main menu
- Available both when authenticated and not authenticated with GitHub

### 2. System Installation Menu
The system installation menu provides:
- Current installation status
- Installation path information
- Sudo requirement detection
- Shell integration status

### 3. Installation Process
The system installer:
- Copies the BOBA binary to `/usr/local/bin/boba` (Linux/macOS) or appropriate Windows location
- Uses `sudo` when required for system directories
- Creates backups of existing installations
- Modifies `~/.zshrc` to add PATH and aliases
- Provides verification and rollback functionality

### 4. Shell Integration
Automatically adds to `~/.zshrc`:
```bash
# BOBA CLI Tool Configuration
# Added by BOBA installer on [timestamp]
export PATH="/usr/local/bin:$PATH"
alias boba-update="boba"
alias dev-setup="boba"

# BOBA completion (if available)
if command -v boba >/dev/null 2>&1; then
    # Add any completion setup here in the future
    :
fi
```

### 5. User Privilege Detection
- Automatically detects if sudo is required
- Shows appropriate warnings in the UI
- Handles Windows elevation differently

### 6. Verification and Rollback
- Verifies binary installation
- Checks shell configuration
- Provides uninstall functionality
- Maintains backups for rollback

## Usage Flow

1. Start BOBA application
2. Navigate to "🔧 Install BOBA to System"
3. Review installation details and requirements
4. Select "▶️ Start System Installation"
5. Follow prompts for sudo if required
6. Verify successful installation
7. Restart shell or run `source ~/.zshrc`
8. Use `boba` command system-wide

## Error Handling

The system installer handles:
- Missing sudo privileges
- File permission errors
- Backup creation failures
- Shell configuration errors
- Installation verification failures

## Testing

Comprehensive tests cover:
- System installer initialization
- Platform detection
- Privilege requirement detection
- Configuration generation
- File operations
- Menu navigation
- UI state management
- Error conditions

## Platform Support

- **Linux**: Installs to `/usr/local/bin`, uses sudo when needed
- **macOS**: Installs to `/usr/local/bin`, uses sudo when needed  
- **Windows**: Installs to user-accessible location, handles UAC appropriately

## Security Considerations

- Only modifies user's shell configuration files
- Creates backups before making changes
- Uses system-standard installation paths
- Proper file permissions (755 for binary, 644 for config)
- Validates installation paths and prevents directory traversal