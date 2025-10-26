# Workflow Dependencies

This document explains how the GitHub Actions workflows are connected and their dependencies.

## 🔄 Workflow Chain

```
┌─────────────────────────────────────────────────────────────┐
│                    Push to main/master                       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │   CI Workflow    │
                    │   (ci.yml)       │
                    └──────────────────┘
                              │
                    ┌─────────┴─────────┐
                    │                   │
                    ▼                   ▼
              ✅ Success          ❌ Failure
                    │                   │
                    ▼                   ▼
          ┌──────────────────┐    Release blocked
          │ Release Workflow │    (no release created)
          │  (release.yml)   │
          └──────────────────┘
                    │
                    ▼
          Development Release
          (v0.0.0-dev-<hash>)


┌─────────────────────────────────────────────────────────────┐
│                    Push version tag (v*.*.*)                 │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │ Tagged Release   │
                    │ Workflow         │
                    │ (tag-release.yml)│
                    └──────────────────┘
                              │
                              ▼
                    Production Release
                    (v1.0.0, v1.1.0, etc.)
```

## 📋 Workflow Details

### 1. CI Workflow (`ci.yml`)

**Triggers:**
- Pull requests to main/master
- Pushes to main/master

**Purpose:**
- Run tests on all platforms (Linux, macOS, Windows)
- Perform linting and code quality checks
- Generate code coverage reports
- Verify the build succeeds

**Dependencies:**
- None (runs independently)

**Blocks:**
- Release workflow (if CI fails, release won't run)

### 2. Release Workflow (`release.yml`)

**Triggers:**
- **After CI workflow succeeds** on main/master
- Direct push of version tags (v*)

**Purpose:**
- Create development releases for testing
- Build multi-platform binaries
- Generate pre-releases

**Dependencies:**
- **Requires CI workflow to succeed** (for branch pushes)
- For tag pushes: runs independently with own tests

**Behavior:**
- For branch pushes: Only runs if CI passed
- For tag pushes: Runs immediately with own test suite

### 3. Tagged Release Workflow (`tag-release.yml`)

**Triggers:**
- Push of version tags matching `v*.*.*`

**Purpose:**
- Create official production releases
- Build optimized binaries
- Generate comprehensive release notes

**Dependencies:**
- None (runs independently)
- Includes its own test suite

## 🎯 Why This Structure?

### CI Must Pass Before Release

**Problem:** We don't want to create releases if tests are failing.

**Solution:** Release workflow waits for CI to succeed before running.

**Benefits:**
- No broken releases
- Tests run on all platforms before release
- Faster feedback on failures

### Tag Releases Are Independent

**Problem:** Sometimes you need to create a release even if the latest CI run failed (e.g., hotfix).

**Solution:** Tag releases run independently with their own tests.

**Benefits:**
- Can create releases on-demand
- Hotfixes can be released quickly
- Full test suite still runs before release

## 🔧 Configuration

### Release Workflow Dependency

The release workflow uses `workflow_run` to wait for CI:

```yaml
on:
  workflow_run:
    workflows: ["CI"]
    types:
      - completed
    branches:
      - main
      - master
  push:
    tags:
      - 'v*'
```

### Check Job

A check job ensures CI passed:

```yaml
jobs:
  check:
    name: Check CI Status
    runs-on: ubuntu-latest
    if: |
      github.event_name == 'push' ||
      (github.event_name == 'workflow_run' && github.event.workflow_run.conclusion == 'success')
```

## 📊 Workflow Scenarios

### Scenario 1: Push to main (CI passes)

```
1. Developer pushes to main
2. CI workflow runs → ✅ Success
3. Release workflow triggers automatically
4. Development release created
```

### Scenario 2: Push to main (CI fails)

```
1. Developer pushes to main
2. CI workflow runs → ❌ Failure
3. Release workflow does NOT run
4. No release created (as expected)
```

### Scenario 3: Create version tag

```
1. Developer creates tag v1.0.0
2. Tagged Release workflow runs immediately
3. Runs its own tests
4. If tests pass → Production release created
5. If tests fail → No release created
```

### Scenario 4: Pull request

```
1. Developer creates PR
2. CI workflow runs → Shows status
3. No release workflow runs (PRs don't trigger releases)
4. Merge only allowed if CI passes (if branch protection enabled)
```

## ⚙️ Customization

### To Make Releases Independent

If you want releases to run regardless of CI status:

```yaml
# In release.yml, remove the workflow_run trigger
on:
  push:
    branches:
      - main
      - master
    tags:
      - 'v*'
```

### To Require Manual Approval

Add a manual approval step:

```yaml
jobs:
  approval:
    runs-on: ubuntu-latest
    environment: production  # Requires approval in GitHub settings
    steps:
      - name: Wait for approval
        run: echo "Approved"
```

### To Add More Checks

Add additional jobs that must pass:

```yaml
jobs:
  security-scan:
    runs-on: ubuntu-latest
    steps:
      - name: Run security scan
        run: # security scanning commands

  release:
    needs: [test, security-scan]  # Wait for both
```

## 🐛 Troubleshooting

### Release Not Triggering

**Problem:** CI passed but release didn't run

**Check:**
1. Verify workflow_run trigger is correct
2. Check workflow permissions (needs `contents: write`)
3. Look at Actions tab for workflow_run events
4. Ensure branch name matches (main vs master)

**Solution:**
```bash
# Check recent workflow runs
gh run list --workflow=release.yml

# View workflow logs
gh run view <run-id> --log
```

### Release Running When CI Failed

**Problem:** Release created despite CI failure

**Check:**
1. Verify the check job condition
2. Ensure workflow_run conclusion check is correct
3. Check if it was a tag push (which bypasses CI check)

**Solution:**
Update the check job condition to be more strict.

### Both Workflows Running on Tag Push

**Problem:** Both release.yml and tag-release.yml run on tags

**Expected:** This is normal! 
- `release.yml` handles general releases
- `tag-release.yml` is specifically for version tags
- They can coexist or you can adjust triggers

**Solution:**
If you want only one, adjust the triggers:

```yaml
# In release.yml, exclude tags
on:
  workflow_run:
    workflows: ["CI"]
    types:
      - completed
    branches:
      - main
      - master
  # Remove the tags trigger
```

## 📚 Best Practices

### 1. Always Run Tests

Every workflow that creates releases should run tests:
- ✅ CI workflow runs tests
- ✅ Release workflow runs tests
- ✅ Tagged release workflow runs tests

### 2. Use Branch Protection

Enable branch protection rules:
- Require CI to pass before merging
- Require pull request reviews
- Prevent direct pushes to main

### 3. Use Environments for Production

For production releases, use GitHub environments:
- Require manual approval
- Add deployment protection rules
- Limit who can approve

### 4. Monitor Workflow Runs

Regularly check:
- Failed workflow runs
- Workflow execution times
- Resource usage

### 5. Test Workflow Changes

Before merging workflow changes:
- Test in a fork or branch
- Use `workflow_dispatch` for manual testing
- Verify with a test tag

## 🔗 Related Documentation

- [CI Workflow](TESTING.md#ci-workflow)
- [Release Process](RELEASE.md)
- [Quick Reference](RELEASE_QUICK_REFERENCE.md)
- [Pipeline Summary](PIPELINE_SUMMARY.md)

## ✅ Summary

- **CI must pass** before development releases are created
- **Tag releases** run independently with their own tests
- **Workflow dependencies** prevent broken releases
- **Flexible structure** allows for hotfixes and emergency releases

This structure ensures quality while maintaining flexibility for different release scenarios.
