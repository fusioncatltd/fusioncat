# GitHub Actions Workflows

This directory contains GitHub Actions workflows for building and publishing Docker images.

## Workflow

### Docker Publish (`docker-publish.yml`)
- **Trigger**: When a version tag is pushed (e.g., `v1.0.0`, `v2.1.3`)
- **Purpose**: Builds and publishes Docker images to GitHub Container Registry (ghcr.io)
- **Features**:
  - Multi-platform builds (linux/amd64, linux/arm64)
  - Semantic versioning tags
  - Automatic `latest` tag
  - GitHub Actions cache for faster builds

## How to Use

### Publishing a New Version

1. Create and push a version tag:
```bash
git tag v1.0.0
git push origin v1.0.0
```

2. The workflow will automatically:
   - Build the Docker image
   - Push to `ghcr.io/fusioncatltd/fusioncat:v1.0.0`
   - Also tag as `latest` if on main branch

### Pulling the Published Image

```bash
# Pull latest version
docker pull ghcr.io/fusioncatltd/fusioncat:latest

# Pull specific version
docker pull ghcr.io/fusioncatltd/fusioncat:v1.0.0

# Run the container
docker run -p 8080:8080 ghcr.io/fusioncatltd/fusioncat:latest
```

## Testing Locally

Before pushing to GitHub, you can test locally:

### Test Docker Build
```bash
# Test Docker build and container startup
make docker-test
```

### Validate Workflow Syntax
```bash
# Validate GitHub Actions workflow syntax
make gh-validate

# Or check with act (requires act installation)
make gh-actions-test
```

### Install act (optional, for workflow syntax checking)
```bash
# macOS
brew install act

# Linux
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Or using Make
make install-act
```

## Configuration

### Required Secrets
No additional secrets are required. The workflows use the built-in `GITHUB_TOKEN` for authentication to GitHub Container Registry.

### Permissions
The repository must have the following permissions enabled:
- **Packages**: Write (for pushing to ghcr.io)
- **Contents**: Read (for checking out code)

These are configured in the workflow files.

## Container Registry

Images are published to GitHub Container Registry (ghcr.io):
- **Registry**: `ghcr.io`
- **Namespace**: `fusioncatltd` (your organization)
- **Image**: `fusioncat`
- **Full path**: `ghcr.io/fusioncatltd/fusioncat`

### Making Images Public

By default, container images are private. To make them public:

1. Go to your organization's packages: https://github.com/orgs/fusioncatltd/packages
2. Find the `fusioncat` package
3. Go to Package Settings
4. Change visibility to Public

## Troubleshooting

### Workflow Not Triggering
- Ensure the tag follows the pattern `v*.*.*` (e.g., v1.0.0)
- Check Actions tab in GitHub for any errors

### Build Failures
- Check the workflow logs in the Actions tab
- Test locally with `make docker-test`
- Verify Dockerfile syntax

### Permission Errors
- Ensure the workflow has the correct permissions (see Configuration section)
- Check that GITHUB_TOKEN has package write access

## Version Tag Format

Use semantic versioning for tags:
- `v1.0.0` - Major.Minor.Patch
- `v1.0.0-beta.1` - Pre-release versions
- `v2.0.0-rc.1` - Release candidates

The workflow will generate the following tags:
- Full version: `v1.0.0`
- Major.Minor: `v1.0`
- Major only: `v1`
- Latest: `latest` (for stable releases)