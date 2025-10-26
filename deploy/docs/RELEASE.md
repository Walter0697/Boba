# Release Process

This document describes the automated release process for BOBA.

## Overview

BOBA uses GitHub Actions to automate the build, test, and release process. There are three main workflows:

1. **CI Workflow** (`ci.yml`) - Runs on every push and pull request
2. **Release Workflow** (`release.yml`) - Creates development releases on main/master branch pushes
3. **Tagged Release Workflow** (`tag-release.yml`) - Creates official releases when version tags are pushed

## Automated Workflows

### Continuous Integration (CI)

**Trigger:** Pull requests and pushes to main/master

**Actions:**
- Runs tests on Linux, macOS, and Windows
- Performs linting with golangci-lint
- Generates code coverage reports
- Builds the binary to verify compilation

### Development Releases

**Trigger:** Pushes to main/master branch (without tags)

**Actions:**
- Runs full test suite
- Builds binaries for all platforms (Linux, macOS, Windows) and architectures (amd64, arm64)
- Creates a pre-release with version `v0.0.0-dev-<commit-hash>`
- Generates release notes from recent commits
- Uploads all platform binaries as release assets

### Official Tagged Releases

**Trigger:** Pushing a version tag (e.g., `v1.0.0`)

**Actions:**
- Runs full test suite
- Builds optimized binaries for all platforms with version information embedded
- Creates compressed archives (tar.gz for Unix, zip for Windows)
- Generates SHA256 checksums for all artifacts
- Creates comprehensive release notes with:
  - Installation instructions for each platform
  - Changelog since last release
  - Checksums for verification
- Publishes official release on GitHub

## Creating a New Release

### 1. Prepare the Release

Ensure all changes are committed and pushed to the main branch:

```bash
git checkout main
git pull origin main
```

### 2. Create and Push a Version Tag

Follow semantic versioning (MAJOR.MINOR.PATCH):

```bash
# For a new feature release
git tag -a v1.1.0 -m "Release v1.1.0: Add new features"

# For a bug fix release
git tag -a v1.0.1 -m "Release v1.0.1: Bug fixes"

# For a major release with breaking changes
git tag -a v2.0.0 -m "Release v2.0.0: Major update"

# Push the tag to trigger the release workflow
git push origin v1.1.0
```

### 3. Monitor the Release

1. Go to the **Actions** tab in your GitHub repository
2. Watch the "Tagged Release" workflow execute
3. Once complete, check the **Releases** page for the new release

### 4. Verify the Release

Download and test the release artifacts:

```bash
# Example for Linux
wget https://github.com/YOUR_USERNAME/boba/releases/download/v1.1.0/boba-linux-amd64.tar.gz
tar -xzf boba-linux-amd64.tar.gz
./boba-linux-amd64 --version
```

## Version Numbering

BOBA follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version (X.0.0): Incompatible API changes or major feature overhauls
- **MINOR** version (0.X.0): New features in a backward-compatible manner
- **PATCH** version (0.0.X): Backward-compatible bug fixes

### Examples

- `v1.0.0` - Initial stable release
- `v1.1.0` - Added new environment setup features
- `v1.1.1` - Fixed bug in GitHub authentication
- `v2.0.0` - Complete UI redesign (breaking changes)

## Build Information

Each release binary includes embedded build information accessible via:

```bash
boba --version
```

This displays:
- Version number (from git tag)
- Build timestamp
- Git commit hash

## Platform Support

Releases include binaries for:

- **Linux**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64

## Troubleshooting

### Release Workflow Fails

1. Check the Actions tab for error details
2. Common issues:
   - Test failures: Fix tests and push changes
   - Build errors: Verify code compiles locally
   - Permission issues: Ensure GitHub Actions has write permissions

### Tag Already Exists

If you need to recreate a tag:

```bash
# Delete local tag
git tag -d v1.0.0

# Delete remote tag
git push origin :refs/tags/v1.0.0

# Create new tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### Missing Release Assets

If release assets are missing:
1. Check the workflow logs for build failures
2. Verify all platforms built successfully
3. Re-run the workflow if needed

## Manual Release (Emergency)

If automated releases fail, you can create a manual release:

```bash
# Build for all platforms
./deploy/scripts/build-all.sh

# Create release manually on GitHub
# Upload the built binaries from the dist/ directory
```

## Release Checklist

Before creating a release:

- [ ] All tests pass locally
- [ ] Documentation is up to date
- [ ] CHANGELOG.md is updated (if you maintain one)
- [ ] Version number follows semantic versioning
- [ ] Breaking changes are documented
- [ ] Dependencies are up to date

After creating a release:

- [ ] Verify release appears on GitHub
- [ ] Test download and installation on at least one platform
- [ ] Announce release (if applicable)
- [ ] Update documentation with new version number
