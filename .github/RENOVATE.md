# Renovate Configuration

This repository uses [Renovate](https://docs.renovatebot.com/) for automated dependency management.

## Configuration Overview

The Renovate configuration (`renovate.json`) is set up to:

1. **Group dependencies by package manager** - All updates for a specific package manager are grouped into a single PR
2. **Use semantic commit messages** - Commits follow the conventional commit format
3. **Schedule updates** - Runs before 6am UTC on Mondays to avoid disrupting the work week
4. **Automerge minor and patch updates** - Small updates are automatically merged after CI passes
5. **Require manual approval for major updates** - Breaking changes need human review
6. **Handle security vulnerabilities** - Security updates are labeled and automerged

## Package Manager Groups

### Go Dependencies
- **Manager**: `gomod`
- **Commit format**: `build(go): <description>`
- **Files**: `go.mod`, `go.sum`

### npm/yarn Dependencies
- **Manager**: `npm`
- **Commit format**: `build(npm): <description>`
- **Files**: `cmd/paretosecurity-tray/ui/package.json`, `cmd/paretosecurity-installer/ui/package.json`

### GitHub Actions
- **Manager**: `github-actions`
- **Commit format**: `ci(actions): <description>`
- **Files**: `.github/workflows/*.yml`

### Python Dependencies
- **Manager**: `pip_requirements`
- **Commit format**: `build(python): <description>`
- **Files**: `.github/workflows/requirements.txt`

## Update Strategy

- **Schedule**: Before 6am UTC on Mondays
- **PR Limits**: Maximum 5 concurrent PRs, 2 per hour
- **Automerge**: Enabled for minor and patch updates
- **Major Updates**: Require manual review and approval

## Labels

All Renovate PRs are automatically labeled with:
- `dependencies` - Indicates this is a dependency update
- `renovate` - Indicates this PR was created by Renovate
- `security` - Added for security vulnerability fixes

## Why Group by Package Manager?

Grouping updates by package manager provides several benefits:

1. **Reduced PR volume** - Instead of individual PRs for each dependency, you get one PR per package manager
2. **Easier review** - All related updates can be reviewed together in context
3. **Better CI efficiency** - Single test run for all updates of the same type
4. **Clearer commit history** - Semantic commits make it easy to understand what changed

## Disabling Renovate

To temporarily disable Renovate, add `"enabled": false` to the `renovate.json` file.

To disable for specific dependencies, add them to `ignoreDeps`:

```json
{
  "ignoreDeps": ["package-name"]
}
```

## More Information

- [Renovate Documentation](https://docs.renovatebot.com/)
- [Configuration Options](https://docs.renovatebot.com/configuration-options/)
- [Package Rules](https://docs.renovatebot.com/configuration-options/#packagerules)
