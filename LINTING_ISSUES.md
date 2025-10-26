# Linting Issues to Fix

This document tracks linting issues found by golangci-lint that should be fixed in future PRs.

## 📋 Current Issues

### 0. Integration Test Failure (CRITICAL)

**Issue:** Integration test failing - stdout not being captured properly.

**Location:**
- `internal/installer/integration_test.go:98` - Expected stdout to be captured

**Root Cause:** The `executeScriptSecurely` function in `engine.go` is not properly capturing stdout/stderr output.

**Impact:** Integration tests fail, CI set to non-blocking temporarily.

**Fix:**

The issue is in `internal/installer/engine.go` where stdout/stderr are being captured. The test expects to see "STDOUT:" in the output, but it's not being captured.

```go
// In engine.go, ensure stdout/stderr are properly captured and included in result
func (e *InstallationEngine) executeScriptSecurely(...) (*InstallationResult, error) {
    var stdout, stderr strings.Builder
    
    // Ensure these are properly synchronized and captured
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    // ... execute command ...
    
    // Include in result
    result.Output = fmt.Sprintf("STDOUT:\n%s\nSTDERR:\n%s", stdout.String(), stderr.String())
    return result, nil
}
```

### 1. Unchecked Error Returns (errcheck)

**Issue:** Error return values are not being checked.

**Locations:**
- `internal/ui/initialization.go:18` - `configManager.LoadConfig`
- `internal/ui/actions.go:243` - `m.configManager.RecordToolInstallation`
- `internal/installer/engine_test.go:191` - `engine.Cleanup`
- `internal/installer/engine_test.go:161` - `engine.Cleanup`
- `internal/installer/engine_test.go:123` - `engine.Cleanup`
- `internal/installer/engine.go:566` - `cmd.Process.Kill`
- `internal/installer/engine.go:313` - `cmd.Process.Kill`
- `internal/installer/engine.go:44` - `os.MkdirAll`

**Fix Examples:**

```go
// Before
configManager.LoadConfig()

// After
if err := configManager.LoadConfig(); err != nil {
    log.Printf("Failed to load config: %v", err)
    // Handle error appropriately
}
```

```go
// Before
cmd.Process.Kill()

// After
if err := cmd.Process.Kill(); err != nil {
    log.Printf("Failed to kill process: %v", err)
}
```

```go
// Before
os.MkdirAll(path, 0755)

// After
if err := os.MkdirAll(path, 0755); err != nil {
    return fmt.Errorf("failed to create directory: %w", err)
}
```

### 2. Unused Fields (unused)

**Issue:** Struct field is declared but never used.

**Location:**
- `internal/github/auth.go:29` - field `cursor` is unused

**Fix:**

```go
// Before
type authModel struct {
    cursor int  // unused
    // other fields...
}

// After - Option 1: Remove if truly unused
type authModel struct {
    // other fields...
}

// After - Option 2: Use it if it should be used
func (m authModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up":
            m.cursor--
        case "down":
            m.cursor++
        }
    }
    return m, nil
}
```

### 3. Ineffectual Assignments (ineffassign)

**Issue:** Variable is assigned but the value is never used.

**Location:**
- `internal/config/manager_test.go:294` - ineffectual assignment to `enabled`

**Fix:**

```go
// Before
enabled := true
// ... code that doesn't use enabled

// After - Option 1: Remove if not needed
// Just remove the line

// After - Option 2: Use the value
enabled := true
if enabled {
    // do something
}

// After - Option 3: Use blank identifier if intentionally ignoring
_ = someFunction() // explicitly ignore return value
```

## 🎯 Priority

### Critical Priority (Blocks Tests)
0. 🔴 Integration test failure (`internal/installer/integration_test.go:98`)
   - Tests are currently set to non-blocking because of this
   - Should be fixed ASAP to re-enable test blocking

### High Priority (Affects Functionality)
1. ✅ Unchecked errors in production code (`internal/ui/`, `internal/installer/engine.go`)
   - These could hide real errors
   - Should be fixed ASAP

### Medium Priority (Affects Tests)
2. ⚠️ Unchecked errors in test code (`internal/installer/engine_test.go`)
   - Less critical but should still be fixed
   - Tests might not catch failures properly

### Low Priority (Code Quality)
3. 📝 Unused fields and ineffectual assignments
   - Doesn't affect functionality
   - Good to clean up for code quality

## 🔧 Recommended Fixes

### For Production Code

Create a PR to fix all production code issues:

```bash
# Create a branch
git checkout -b fix/linting-errors

# Fix the issues in:
# - internal/ui/initialization.go
# - internal/ui/actions.go
# - internal/installer/engine.go
# - internal/github/auth.go

# Test the changes
go test ./...

# Run linter
golangci-lint run

# Commit and push
git add .
git commit -m "fix: address linting errors in production code"
git push origin fix/linting-errors
```

### For Test Code

Create a separate PR for test code:

```bash
# Create a branch
git checkout -b fix/test-linting-errors

# Fix the issues in:
# - internal/installer/engine_test.go
# - internal/config/manager_test.go

# Run tests
go test ./...

# Commit and push
git add .
git commit -m "fix: address linting errors in test code"
git push origin fix/test-linting-errors
```

## 📝 Detailed Fix Guide

### 1. internal/ui/initialization.go:18

```go
// Current code (line 18)
configManager.LoadConfig()

// Fixed code
if err := configManager.LoadConfig(); err != nil {
    // Log the error but don't fail initialization
    // The app can still run with default config
    log.Printf("Warning: Failed to load config: %v", err)
}
```

### 2. internal/ui/actions.go:243

```go
// Current code (line 243)
m.configManager.RecordToolInstallation(tool.Name, version, "auto")

// Fixed code
if err := m.configManager.RecordToolInstallation(tool.Name, version, "auto"); err != nil {
    // Log but don't fail the installation
    log.Printf("Warning: Failed to record installation: %v", err)
}
```

### 3. internal/installer/engine.go:44

```go
// Current code (line 44)
os.MkdirAll(e.workDir, 0755)

// Fixed code
if err := os.MkdirAll(e.workDir, 0755); err != nil {
    return nil, fmt.Errorf("failed to create work directory: %w", err)
}
```

### 4. internal/installer/engine.go:313 & 566

```go
// Current code (lines 313, 566)
cmd.Process.Kill()

// Fixed code
if err := cmd.Process.Kill(); err != nil {
    // Process might already be dead, log but don't fail
    log.Printf("Warning: Failed to kill process: %v", err)
}
```

### 5. internal/installer/engine_test.go (multiple locations)

```go
// Current code
engine.Cleanup()

// Fixed code
if err := engine.Cleanup(); err != nil {
    t.Errorf("Cleanup failed: %v", err)
}

// Or if cleanup failure is not critical
_ = engine.Cleanup() // explicitly ignore error
```

### 6. internal/github/auth.go:29

```go
// Current code
type authModel struct {
    cursor int  // unused
    // ...
}

// Fixed code - remove if truly unused
type authModel struct {
    // ... other fields only
}
```

### 7. internal/config/manager_test.go:294

```go
// Current code (line 294)
enabled := true
// ... code that doesn't use enabled

// Fixed code - either use it or remove it
// Option 1: Remove
// Just delete the line

// Option 2: Use it
enabled := true
if !enabled {
    t.Skip("Test disabled")
}
```

## 🧪 Testing After Fixes

After fixing each issue, run:

```bash
# Run tests
go test ./...

# Run linter
golangci-lint run

# Run specific linters
golangci-lint run --enable=errcheck,unused,ineffassign

# Check specific files
golangci-lint run internal/ui/initialization.go
```

## 📊 Progress Tracking

### Critical
- [ ] Fix `internal/installer/integration_test.go:98` - stdout capture issue

### High Priority
- [ ] Fix `internal/ui/initialization.go:18`
- [ ] Fix `internal/ui/actions.go:243`
- [ ] Fix `internal/installer/engine.go:44`
- [ ] Fix `internal/installer/engine.go:313`
- [ ] Fix `internal/installer/engine.go:566`
- [ ] Fix `internal/installer/engine_test.go:123`
- [ ] Fix `internal/installer/engine_test.go:161`
- [ ] Fix `internal/installer/engine_test.go:191`
- [ ] Fix `internal/github/auth.go:29`
- [ ] Fix `internal/config/manager_test.go:294`

## 🔗 Related

- [Workflow Fixes](WORKFLOW_FIXES.md)
- [CI Configuration](.github/workflows/ci.yml)
- [golangci-lint Documentation](https://golangci-lint.run/)

## ✅ When Complete

Once all issues are fixed:

1. Update this document to mark items as complete
2. Run full test suite: `go test -v ./...`
3. Run linter: `golangci-lint run`
4. Verify CI passes
5. Archive this document or move to `docs/resolved/`

---

**Note:** These issues don't block releases (linting is set to `continue-on-error: true`), but they should be fixed for code quality and to catch potential bugs.
