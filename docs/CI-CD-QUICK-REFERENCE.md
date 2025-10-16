# CI/CD Pipeline Quick Reference

## 🚀 Quick Start

### For Developers

**Running Tests Locally**:
```bash
# Run unit tests
go test -v ./...

# Run with coverage
go test -v -coverprofile=coverage.out ./...

# Run integration tests (requires Docker)
docker-compose -f docker-compose.test.yml up -d
go test -v ./tests/integration/...
docker-compose -f docker-compose.test.yml down
```

**Building Containers Locally**:
```bash
# Build single component
docker build -f docker/Dockerfile.web-dashboard -t agentaflow:web-dashboard .

# Build all components
./docker/build.sh    # Linux/Mac
./docker/build.ps1   # Windows
```

### For Pull Requests

**What Runs Automatically**:
1. ✅ Security scan (Trivy on source code)
2. ✅ Unit tests with Go 1.21
3. ✅ Build validation (all 3 components)
4. ✅ Coverage reporting

**What Doesn't Run**:
- ❌ Container publishing (only on main/develop)
- ❌ Integration tests (only on main/develop)

**Viewing Results**:
- Check the "Checks" tab on your PR
- View detailed logs in Actions tab
- See coverage in PR comments (if Codecov configured)

### For Main Branch

**What Runs on Push to Main**:
1. ✅ Security scan
2. ✅ Unit tests + coverage
3. ✅ Multi-arch container builds
4. ✅ Publish to GitHub Packages
5. ✅ Container security scanning
6. ✅ Integration testing
7. ✅ Build summary

**Published Images**:
```
ghcr.io/finoptimize/agentaflow-sro-community:web-dashboard-latest
ghcr.io/finoptimize/agentaflow-sro-community:web-dashboard-main-<commit-sha>
ghcr.io/finoptimize/agentaflow-sro-community:k8s-scheduler-latest
ghcr.io/finoptimize/agentaflow-sro-community:prometheus-demo-latest
```

### For Release Tags

**Creating a Release**:
```bash
# Tag the release
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

**What Runs**:
- All main branch steps +
- ✅ SLSA provenance generation
- ✅ Semantic version tags

**Published Images**:
```
ghcr.io/finoptimize/agentaflow-sro-community:web-dashboard-v1.0.0
ghcr.io/finoptimize/agentaflow-sro-community:web-dashboard-1.0
ghcr.io/finoptimize/agentaflow-sro-community:web-dashboard-latest
```

---

## 🛠️ Manual Workflow Dispatch

### Running Workflows Manually

1. Go to [Actions tab](https://github.com/Finoptimize/agentaflow-sro-community/actions)
2. Select workflow (Container Build or Security Scan)
3. Click "Run workflow"
4. Choose branch
5. Click green "Run workflow" button

**Use Cases**:
- Testing workflow changes
- Re-running failed builds
- Running security scans on demand
- Testing specific branches

---

## 🔒 Security Scanning

### Automated Scans

**Schedule**: Every Monday at 8:00 AM UTC

**What Gets Scanned**:
1. **Source Code** (Trivy)
   - Vulnerability detection in Go code
   - Configuration issues
   - SARIF upload to GitHub Security

2. **Dependencies** (govulncheck + Nancy)
   - Go module vulnerabilities
   - OSS Index checks
   - Third-party package issues

3. **Containers** (Trivy + Grype)
   - Base image vulnerabilities
   - Layer-by-layer analysis
   - Multi-scanner validation

4. **Code Quality** (CodeQL)
   - Security patterns
   - Code quality issues
   - Extended security queries

5. **Secrets** (Gitleaks)
   - Exposed credentials
   - API keys
   - Private tokens

### Viewing Security Results

**GitHub Security Tab**:
- Navigate to: `https://github.com/Finoptimize/agentaflow-sro-community/security`
- View: Code scanning alerts, Dependabot alerts, Secret scanning

**Action Logs**:
- Navigate to: Actions → Security Scan workflow
- View: Detailed scan results and logs

---

## 📦 Working with Published Images

### Pulling Images

**Authentication**:
```bash
# Login to GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

**Pulling**:
```bash
# Pull latest
docker pull ghcr.io/finoptimize/agentaflow-sro-community:web-dashboard-latest

# Pull specific version
docker pull ghcr.io/finoptimize/agentaflow-sro-community:web-dashboard-v1.0.0

# Pull for specific architecture
docker pull --platform linux/amd64 ghcr.io/finoptimize/agentaflow-sro-community:web-dashboard-latest
```

### Using in Docker Compose

**Update docker-compose.yml**:
```yaml
services:
  agentaflow-dashboard:
    image: ghcr.io/finoptimize/agentaflow-sro-community:web-dashboard-v1.0.0
    # ... rest of config
```

**Pull and restart**:
```bash
docker-compose pull
docker-compose up -d
```

### Using in Kubernetes

**Deployment manifest**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agentaflow-dashboard
spec:
  template:
    spec:
      containers:
      - name: dashboard
        image: ghcr.io/finoptimize/agentaflow-sro-community:web-dashboard-v1.0.0
        imagePullPolicy: IfNotPresent
```

**Apply**:
```bash
kubectl apply -f deployment.yaml
```

---

## 🤖 Dependabot

### How It Works

**Schedule**:
- **Go modules**: Mondays 9:00 AM UTC
- **Docker images**: Mondays 10:00 AM UTC  
- **GitHub Actions**: Mondays 11:00 AM UTC

**What Happens**:
1. Dependabot checks for updates
2. Creates PR with dependency update
3. CI runs tests on PR
4. Auto-merge if tests pass (configure in settings)

### Managing Dependabot PRs

**Reviewing**:
```bash
# View Dependabot PRs
gh pr list --label dependencies

# Review specific PR
gh pr view <PR-NUMBER>
```

**Commands** (comment on PR):
- `@dependabot rebase` - Rebase the PR
- `@dependabot recreate` - Recreate the PR
- `@dependabot merge` - Merge the PR
- `@dependabot close` - Close the PR
- `@dependabot ignore this dependency` - Skip this dependency

---

## 🐛 Troubleshooting

### Build Failures

**Check logs**:
1. Go to Actions tab
2. Click failed workflow run
3. Click failed job
4. Expand failed step

**Common Issues**:

**Issue**: Go tests fail
```bash
# Run locally to reproduce
go test -v ./...
```

**Issue**: Docker build fails
```bash
# Check Dockerfile syntax
docker build -f docker/Dockerfile.web-dashboard --no-cache .
```

**Issue**: Integration tests fail
```bash
# Run integration tests locally
docker-compose -f docker-compose.test.yml up -d
docker-compose -f docker-compose.test.yml logs
go test -v ./tests/integration/...
```

### Publishing Failures

**Issue**: Permission denied pushing to ghcr.io

**Solution**: Ensure `GITHUB_TOKEN` has `packages: write` permission
- Check workflow file has `permissions: packages: write`
- Verify repository settings allow package publishing

**Issue**: Multi-arch build fails

**Solution**: Check Buildx setup
```bash
# Test locally
docker buildx create --use
docker buildx build --platform linux/amd64,linux/arm64 .
```

### Security Scan Failures

**Issue**: Too many vulnerabilities detected

**Solution**: 
1. Review in Security tab
2. Update dependencies: `go get -u all`
3. Check base image updates in Dockerfiles
4. Create issues for legitimate findings

---

## 📊 Monitoring CI/CD

### Key Metrics to Watch

**Build Health**:
- ✅ Passing builds on main
- ✅ Green status badges
- ✅ Integration tests passing

**Security Posture**:
- 🔒 Zero critical vulnerabilities
- 🔒 Dependabot PRs merged within 7 days
- 🔒 All security scans passing

**Performance**:
- ⚡ Build time < 15 minutes
- ⚡ Cache hit rate > 80%
- ⚡ Integration tests < 2 minutes

### Dashboards

**GitHub Actions**:
- View: https://github.com/Finoptimize/agentaflow-sro-community/actions
- Metrics: Build duration, success rate, workflow runs

**GitHub Insights**:
- View: https://github.com/Finoptimize/agentaflow-sro-community/pulse
- Metrics: Commits, PRs, contributors

**GitHub Packages**:
- View: https://github.com/Finoptimize/agentaflow-sro-community/pkgs/container/agentaflow-sro-community
- Metrics: Downloads, versions, storage

---

## 🎯 Best Practices

### For Contributors

1. **Run tests locally** before pushing
2. **Keep commits small** and focused
3. **Wait for CI checks** before requesting review
4. **Fix security issues** promptly
5. **Update documentation** with code changes

### For Maintainers

1. **Review Dependabot PRs weekly**
2. **Monitor security scans**
3. **Update base images** regularly
4. **Tag releases** with semantic versioning
5. **Keep workflows up to date**

### For Releases

1. **Update CHANGELOG.md**
2. **Test thoroughly** on develop branch
3. **Tag with semantic version**: `v1.0.0`
4. **Create GitHub Release** with notes
5. **Announce** in README and docs

---

## 📚 Additional Resources

**GitHub Actions**:
- [Workflow syntax](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)
- [Workflow commands](https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions)

**GitHub Packages**:
- [Publishing packages](https://docs.github.com/en/packages/learn-github-packages/publishing-a-package)
- [Installing packages](https://docs.github.com/en/packages/learn-github-packages/installing-a-package)

**Security**:
- [Code scanning](https://docs.github.com/en/code-security/code-scanning)
- [Dependabot](https://docs.github.com/en/code-security/dependabot)
- [Secret scanning](https://docs.github.com/en/code-security/secret-scanning)

**Docker**:
- [Multi-arch builds](https://docs.docker.com/build/building/multi-platform/)
- [Build cache](https://docs.docker.com/build/cache/)
- [GitHub Actions cache](https://github.com/docker/build-push-action/blob/master/docs/advanced/cache.md)

---

## 🆘 Getting Help

**Issues**:
- Create issue: https://github.com/Finoptimize/agentaflow-sro-community/issues/new
- Label: `ci/cd`, `bug`, `documentation`

**Discussions**:
- Ask questions: https://github.com/Finoptimize/agentaflow-sro-community/discussions
- Share feedback: `Ideas` category

**Documentation**:
- Main README: [README.md](../README.md)
- Docker docs: [docker/README.md](../docker/README.md)
- Container strategy: [CONTAINER.md](../CONTAINER.md)
- Phase 2 details: [PHASE2-COMPLETE.md](PHASE2-COMPLETE.md)
