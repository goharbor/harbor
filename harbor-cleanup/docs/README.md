# Harbor Cleanup Tool

A comprehensive Python tool for cleaning up Harbor registry artifacts based on configurable retention policies.

## 🏗️ Architecture

This tool features a modular architecture with proper Python package organization:

```
harbor-cleanup/
├── harbor_cleanup/              # 📦 Main package
│   ├── core/                   # 🔧 Core business logic
│   │   ├── config.py           # Configuration management
│   │   ├── processor.py        # Repository processing logic
│   │   └── retention_policy.py # Retention policy rules
│   ├── api/                    # 🌐 External interfaces
│   │   └── harbor_api.py       # Harbor API client
│   └── utils/                  # 🛠️ Common utilities
│       ├── formatting.py       # Format & utility functions
│       └── reporting.py        # Output & summary generation
├── main.py                     # 🚀 Entry point
├── tests/                      # 🧪 Test modules
├── docs/                       # 📚 Documentation
└── scripts/                    # 📜 Helper scripts
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed information about the modular design.

## 💬 Slack Integration

The tool can send rich Slack notifications for cleanup operations, providing comprehensive summaries of:

### 🚀 **Start Notifications**
- Cleanup mode (Dry Run vs Live Run)
- Target project and repositories
- Start timestamp

### ✅ **Completion Notifications**
- **Repository Summary**: Total processed, successful, failed
- **Artifact Summary**: Total processed, deleted, kept, cleanup percentage
- **Storage Summary**: Total storage, space to cleanup, space to remain
- **Timing**: Duration and completion status

### ❌ **Error Notifications**
- Critical errors and failures
- Interruption notifications (Ctrl+C)
- Detailed error messages

### **Setup Instructions**

1. **Create a Slack App:**
   - Go to [Slack API](https://api.slack.com/apps)
   - Create a new app for your workspace
   - Add `chat:write` OAuth scope
   - Install the app to your workspace

2. **Get Required Values:**
   ```bash
   # Bot User OAuth Token (starts with xoxb-)
   export SLACK_BOT_TOKEN="xoxb-your-slack-bot-token"
   
   # Channel ID (C1234567890) or channel name (#harbor-cleanup)
   export SLACK_CHANNEL_ID="C1234567890"
   
   # Enable notifications
   export SLACK_ENABLED=true
   ```

3. **Add Bot to Channel:**
   ```bash
   # IMPORTANT: Invite the bot to your target channel
   # In Slack, go to your channel and type:
   /invite @your-bot-name
   ```

4. **Test Notifications:**
   ```bash
   # Run a dry-run to test Slack integration
   python main.py
   ```

### **🔧 Troubleshooting**

#### `❌ Slack API error: not_in_channel`
**Solution:** The bot needs to be added to the target channel.
- Go to your Slack channel
- Type `/invite @your-bot-name` 
- Or manually add the bot via channel settings

#### `❌ Slack API error: invalid_auth`
**Solution:** Check your bot token.
- Ensure `SLACK_BOT_TOKEN` starts with `xoxb-`
- Regenerate token if needed in Slack App settings

#### `❌ Slack API error: channel_not_found`
**Solution:** Check your channel ID or name.
- Use channel ID (C1234567890) for private channels
- Use `#channel-name` for public channels
- Get channel ID from channel settings → About

#### `⚠️ Slack notifications enabled but SLACK_BOT_TOKEN not provided`
**Solution:** Set the required environment variables.
```bash
export SLACK_BOT_TOKEN="xoxb-your-token"
export SLACK_CHANNEL_ID="your-channel-id"
```

## 🚀 Quick Start

### Installation

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd harbor-cleanup
   ```

2. **Install dependencies:**
   ```bash
   pip install -r requirements.txt
   ```

3. **Set up environment variables:**
   ```bash
   export HARBOR_USERNAME="your_username"
   export HARBOR_PASSWORD="your_password"
   export HARBOR_PROJECT="your_project"
   export DRY_RUN=true  # Start with dry run for safety
   
   # Optional: Slack notifications
   export SLACK_ENABLED=true
   export SLACK_BOT_TOKEN="xoxb-your-slack-bot-token"
   export SLACK_CHANNEL_ID="C1234567890"  # or "#channel-name"
   ```

### Basic Usage

```bash
# Run dry-run mode (safe - no actual deletions)
python main.py

# Run specific repositories
export REPO_LIST="authz,api,ledger"
export USE_REPO_LIST=true
python main.py

# Run as a Python module
python -m harbor_cleanup

# Run tests to verify installation
PYTHONPATH=. python tests/test_structure.py
```

### Using Docker

```bash
# Build the image
docker build -f docker/Dockerfile -t harbor-cleanup .

# Run with environment variables
docker run --rm \
  -e HARBOR_USERNAME="$HARBOR_USERNAME" \
  -e HARBOR_PASSWORD="$HARBOR_PASSWORD" \
  -e DRY_RUN=true \
  harbor-cleanup
```

## Features

- 🏷️ **Label-based categorization** (prod, non-prod, no-label, other)
- ⏰ **Age-based retention policies** (configurable days)
- 🔢 **Count-based retention policies** (keep last N artifacts)
- 🚀 **Parallel processing** for multiple repositories and pages
- 🔍 **Dry run mode** for safe testing
- 📊 **Detailed reporting** with space usage analysis
- 🗑️ **Automatic garbage collection** triggering
- 🛡️ **Security hardened** with credential validation
- 💬 **Slack notifications** for cleanup start/end summaries
- 📦 **Cross-category size-based deletion ordering** (largest artifacts across all categories deleted first for maximum space cleanup)

## Harbor Capacity and Parallelism Considerations

### Current Parallelism Levels

The tool uses two levels of parallelism:

1. **Repository-level parallelism** (`MAX_WORKERS`): Number of repositories processed simultaneously
2. **Page-level parallelism** (`MAX_PAGE_WORKERS`): Number of concurrent page fetches per repository

### Recommended Settings by Harbor Environment

| Environment | MAX_WORKERS | MAX_PAGE_WORKERS | Max Concurrent Requests | Notes |
|-------------|-------------|------------------|------------------------|-------|
| **Small Harbor** | 2 | 2 | 4 | < 1000 repositories |
| **Medium Harbor** | 5 | 3 | 15 | 1000-10000 repositories |
| **Large Harbor** | 8 | 3 | 24 | > 10000 repositories |
| **Enterprise Harbor** | 10 | 5 | 50 | High-performance setup |

### Harbor API Limits

Most Harbor installations have these typical limits:
- **API rate limit**: 10-100 requests/second
- **Concurrent connections**: 20-50 simultaneous connections
- **Database connections**: 50-200 connections

### Calculating Safe Limits

**Maximum concurrent API requests** = `MAX_WORKERS × MAX_PAGE_WORKERS`

**Examples:**
- Conservative: `MAX_WORKERS=5, MAX_PAGE_WORKERS=2` → 10 concurrent requests
- Moderate: `MAX_WORKERS=5, MAX_PAGE_WORKERS=3` → 15 concurrent requests  
- Aggressive: `MAX_WORKERS=10, MAX_PAGE_WORKERS=5` → 50 concurrent requests

### Signs Harbor is Overwhelmed

Watch for these warning signs:
- API timeout errors increasing
- Retry attempts frequently triggered
- Harbor UI becoming slow/unresponsive
- Database connection errors in Harbor logs
- High CPU/memory usage on Harbor server

### Tuning Recommendations

1. **Start Conservative**: Begin with lower values and increase gradually
2. **Monitor Harbor**: Watch CPU, memory, and response times
3. **Use Retries**: The built-in retry logic helps handle temporary overload
4. **Peak Hours**: Reduce parallelism during business hours
5. **Large Cleanups**: Consider running overnight with higher parallelism

### Emergency Throttling

If Harbor becomes overwhelmed:
```bash
export MAX_WORKERS=1
export MAX_PAGE_WORKERS=1
export API_RETRY_DELAY=2.0
```

This reduces to sequential processing with longer delays between retries.

## Quick Start

### Local Development

1. **Setup environment:**
   ```bash
   cd tools/harbor-cleanup
   cp setup_local_env.sh setup_local_env.sh.local
   # Edit setup_local_env.sh.local with your credentials
   source setup_local_env.sh.local
   ```

2. **Install dependencies:**
   ```bash
   pip install -r requirements.txt
   ```

3. **Run in dry-run mode:**
   ```bash
   python3 main.py
   ```

### Docker Usage

1. **Build the image:**
   ```bash
   docker build -t harbor-cleanup .
   ```

2. **Run with environment variables:**
   ```bash
   docker run --rm \
     -e HARBOR_URL="https://harbor.razorpay.com" \
     -e HARBOR_PROJECT="razorpay" \
     -e HARBOR_USERNAME="your-username" \
     -e HARBOR_PASSWORD="your-password" \
     -e DRY_RUN=true \
     -e LOG_LEVEL=INFO \
     harbor-cleanup
   ```

3. **Use pre-built image from registry:**
   ```bash
   docker run --rm \
     -e HARBOR_URL="https://harbor.razorpay.com" \
     -e HARBOR_PROJECT="razorpay" \
     -e HARBOR_USERNAME="your-username" \
     -e HARBOR_PASSWORD="your-password" \
     c.rzp.io/razorpay/harbor-cleanup:latest
   ```

## Configuration

### Required Environment Variables

```bash
HARBOR_URL           # Harbor server URL (default: https://harbor.razorpay.com)
HARBOR_PROJECT       # Harbor project name (default: razorpay)
HARBOR_USERNAME      # Harbor username (required)
HARBOR_PASSWORD      # Harbor password (required)
```

### Retention Policies

```bash
# Age-based policies (days)
PROD_LABEL_CLEANUP_DAYS=45          # Prod artifacts older than 45 days
NON_PROD_LABEL_CLEANUP_DAYS=14      # Non-prod artifacts older than 14 days
EMPTY_LABEL_CLEANUP_DAYS=45         # No-label artifacts older than 45 days
OTHER_LABEL_CLEANUP_DAYS=45         # Other labeled artifacts older than 45 days

# Count-based policies (number of artifacts to retain regardless of age)
RETAIN_LAST_PROD=100                # Keep last 100 prod artifacts
RETAIN_LAST_NON_PROD=0              # Keep last 0 non-prod artifacts
RETAIN_LAST_NO_LABEL=500            # Keep last 500 no-label artifacts
RETAIN_LAST_OTHER=0                 # Keep last 0 other artifacts
```

### Repository Selection

```bash
# Process specific repositories
USE_REPO_LIST=true                  # Enable repository list mode
REPO_LIST="authz,api,ledger"        # Comma-separated repository names

# Process all repositories (default)
USE_REPO_LIST=false                 # Process all repositories in project
```

### Performance & Behavior

```bash
# Parallel processing
ENABLE_PARALLEL=true                # Enable parallel processing
MAX_WORKERS=5                       # Maximum parallel workers
PARALLEL_PAGE_THRESHOLD=500         # Use parallel pages for repos with >N artifacts
MIN_ARTIFACTS_THRESHOLD=100         # Skip repos with <N artifacts

# Logging
LOG_LEVEL=INFO                      # DEBUG, INFO, WARNING, ERROR
VERBOSE_ARTIFACTS=false             # Show detailed artifact logs

# Safety
DRY_RUN=true                        # Set to false to actually delete artifacts
TRIGGER_GC_AFTER_CLEANUP=false     # Auto-trigger garbage collection
```

## CI/CD Pipeline

The tool includes a comprehensive CI/CD pipeline that:

### On Pull Requests:
- ✅ **Code quality checks** (linting, syntax validation)
- 🔒 **Security scanning** (hardcoded secrets detection)
- 🐳 **Docker image building** and testing
- 🧪 **Dry-run testing** with the built image

### On Push to main/master:
- 🚀 **Automated Docker image building**
- 📦 **Multi-platform image publishing** to Harbor registry
- 📋 **SBOM generation** for security compliance
- 🏷️ **Automated tagging** (latest, branch, SHA)

### Manual Triggers:
- 🔧 **Workflow dispatch** for manual builds
- 🎯 **Force build and push** option

### Pipeline Stages:

1. **Test Stage**
   ```yaml
   - Python syntax validation
   - Code linting with pylint
   - Shell script validation
   - Dependency installation test
   ```

2. **Security Stage**
   ```yaml
   - Hardcoded secrets detection
   - Credential validation
   - Security best practices check
   ```

3. **Build Stage**
   ```yaml
   - Multi-stage Docker build
   - Harbor registry push
   - SBOM generation
   - Artifact upload
   ```

4. **Integration Test Stage**
   ```yaml
   - Dry-run container test
   - Environment validation
   - Configuration testing
   ```

## Docker Image Details

### Base Image
- **Runtime**: `python:3.11-slim-bullseye`
- **Registry**: `c.rzp.io/proxy_dockerhub/library/`
- **Security**: Non-root user (`harborcleaner`)

### Image Features
- 🏗️ **Multi-stage build** for smaller final image
- 🔒 **Security hardened** (non-root user, minimal dependencies)
- 🩺 **Health checks** included
- 📦 **Optimized layers** for better caching
- 🌍 **Environment variables** for configuration

### Image Tags
- `latest` - Latest stable release from main branch
- `main-<sha>` - Builds from main branch with commit SHA
- `pr-<number>` - Pull request builds
- `<branch>-<sha>` - Feature branch builds

## Usage Examples

### Dry Run (Safe Mode)
```bash
# Test cleanup on specific repositories with Slack notifications
docker run --rm \
  -e HARBOR_USERNAME="$HARBOR_USERNAME" \
  -e HARBOR_PASSWORD="$HARBOR_PASSWORD" \
  -e DRY_RUN=true \
  -e USE_REPO_LIST=true \
  -e REPO_LIST="authz,api" \
  -e LOG_LEVEL=INFO \
  -e SLACK_ENABLED=true \
  -e SLACK_BOT_TOKEN="$SLACK_BOT_TOKEN" \
  -e SLACK_CHANNEL_ID="$SLACK_CHANNEL_ID" \
  c.rzp.io/razorpay/harbor-cleanup:latest
```

### Production Cleanup
```bash
# DANGEROUS: Actually delete artifacts
docker run --rm \
  -e HARBOR_USERNAME="$HARBOR_USERNAME" \
  -e HARBOR_PASSWORD="$HARBOR_PASSWORD" \
  -e DRY_RUN=false \
  -e TRIGGER_GC_AFTER_CLEANUP=true \
  -e MIN_ARTIFACTS_THRESHOLD=1000 \
  c.rzp.io/razorpay/harbor-cleanup:latest
```

### Custom Retention Policies
```bash
# Aggressive cleanup for non-prod environments
docker run --rm \
  -e HARBOR_USERNAME="$HARBOR_USERNAME" \
  -e HARBOR_PASSWORD="$HARBOR_PASSWORD" \
  -e NON_PROD_LABEL_CLEANUP_DAYS=7 \
  -e RETAIN_LAST_NON_PROD=10 \
  -e EMPTY_LABEL_CLEANUP_DAYS=14 \
  c.rzp.io/razorpay/harbor-cleanup:latest
```

## Security Considerations

- 🔒 **Never commit credentials** to version control
- 🛡️ **Use environment variables** for sensitive data
- 🧪 **Always test in dry-run mode** first
- 📊 **Review cleanup reports** before live runs
- 🎯 **Use minimal artifact thresholds** for safety
- 🗂️ **Limit repository scope** when possible

## Monitoring & Logging

The tool provides comprehensive logging with configurable verbosity:

- **INFO**: High-level progress and summaries
- **DEBUG**: Detailed artifact-by-artifact processing
- **WARNING**: Non-fatal issues and fallbacks
- **ERROR**: Critical failures requiring attention

### Log Analysis
```bash
# Monitor cleanup progress
docker logs -f <container-id>

# Extract summary statistics
docker logs <container-id> | grep "OVERALL SUMMARY" -A 20

# Find processing times
docker logs <container-id> | grep "processed in"
```

## Contributing

1. Make changes to the tool
2. Update tests and documentation
3. Ensure CI/CD pipeline passes
4. Test with dry-run mode thoroughly
5. Submit pull request

The CI/CD pipeline will automatically validate code quality, security, and functionality. 