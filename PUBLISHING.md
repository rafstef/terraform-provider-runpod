# RunPod Terraform Provider Repository Setup

## Repository Location
`/Users/books/repos/terraform-provider-runpod`

## What's Included
- Complete Terraform provider code for RunPod API
- Darwin ARM64 binary (`terraform-provider-runpod`)
- All resource and data source implementations
- Example configurations
- Provider specification (OpenAPI-based)

## Next Steps to Publish

### 1. Push to GitHub (requires SAML SSO setup)
```bash
cd /Users/books/repos/terraform-provider-runpod
git push -u origin main
```

**Note:** The `runpod` GitHub organization requires SAML SSO. You'll need to:
- Use HTTPS with a personal access token, OR
- Use SSH with an SSH key authorized for the organization

See: https://docs.github.com/articles/authenticating-to-a-github-organization-with-saml-single-sign-on/

### 2. Publish to Terraform Registry
Once the repo is on GitHub, follow the [Terraform Provider Publishing Guide](https://developer.hashicorp.com/terraform/registry/publishing):

```bash
# Tag a release
git tag v1.0.0
git push --tags

# Then publish via Terraform CLI or registry web UI
```

### 3. Build binaries for all platforms
To create releases with binaries for Windows, Linux, and macOS:

```bash
# Darwin (already done)
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o terraform-provider-runpod_darwin_arm64

# Linux AMD64
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o terraform-provider-runpod_linux_amd64

# Linux ARM64  
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o terraform-provider-runpod_linux_arm64

# Windows AMD64
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o terraform-provider-runpod_windows_amd64.exe
```

## Provider Configuration
Users can configure the provider via:
1. Environment variable: `RUNPOD_API_KEY`
2. Provider block in Terraform config

## API Documentation
- REST API: https://rest.runpod.io/v1/docs
- GraphQL API: https://api.runpod.io/graphql
- Docs: https://docs.runpod.io
