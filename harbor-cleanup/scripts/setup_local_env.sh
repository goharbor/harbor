#!/bin/bash

# ⚠️  SECURITY NOTICE:
# 1. Replace XXXXXXXX with your actual Harbor credentials
# 2. Never commit real credentials to version control
# 3. Consider using a .env file and adding it to .gitignore
# 4. This tool can delete artifacts - always test in dry run mode first

# Harbor connection details (replace with your own)
export HARBOR_URL='https://harbor.razorpay.com'    # Harbor server URL
export HARBOR_PROJECT='razorpay'                   # Harbor project name
export HARBOR_USERNAME='XXXXXXXX'                          # Your Harbor username
export HARBOR_PASSWORD='XXXXXXXX'                          # Your Harbor password

# Cleanup policies (days)
export PROD_LABEL_CLEANUP_DAYS=45
export NON_PROD_LABEL_CLEANUP_DAYS=14
export EMPTY_LABEL_CLEANUP_DAYS=45
export OTHER_LABEL_CLEANUP_DAYS=45    # New: separate policy for 'other' category

# Additional label mappings (optional - extends default mappings)
# Default prod labels: prod, production, release, stable, main, master
# Default non-prod labels: nonprod, non-prod, staging, stage, dev, development, test, testing, qa, uat, beta, alpha, feature, develop
# Format: comma-separated list for each category
export ADDITIONAL_PROD_LABELS=""      # e.g., "v1.0,live"
export ADDITIONAL_NONPROD_LABELS=""   # e.g., "sandbox,integration"

# Repository selection
export REPO_LIST="authz,ledger"        # Comma-separated list of repositories
export USE_REPO_LIST=true              # Set to false to process ALL repositories

# Retention policies (number of artifacts to retain regardless of age)
export RETAIN_LAST_PROD=100            # Retain last N prod artifacts
export RETAIN_LAST_NON_PROD=0          # Retain last N non-prod artifacts (0 = disabled)
export RETAIN_LAST_NO_LABEL=500        # Retain last N artifacts with no labels
export RETAIN_LAST_OTHER=0             # Retain last N artifacts with other labels (0 = disabled)

# Parallel processing configuration
export ENABLE_PARALLEL=true            # Enable/disable parallel processing
export MAX_WORKERS=5                   # Maximum number of parallel repository workers
export MAX_PAGE_WORKERS=10             # Maximum number of parallel page fetches per repository
export PARALLEL_PAGE_THRESHOLD=500     # Use parallel page fetching for repos with >N artifacts
export MIN_ARTIFACTS_THRESHOLD=100     # Skip repositories with fewer than N artifacts
export MAX_ARTIFACT_WORKERS=5          # Max concurrent artifact deletions per repository

# Logging configuration
export LOG_LEVEL=INFO                  # DEBUG, INFO, WARNING, ERROR
export VERBOSE_ARTIFACTS=false         # Show detailed artifact-by-artifact output
export PERFORMANCE_LOGGING=true        # Enable performance logging

# Dry run configuration
export DRY_RUN=true                    # Set to false to actually delete artifacts (DANGEROUS!)

# Garbage collection configuration
export TRIGGER_GC_AFTER_CLEANUP=true # Set to true to automatically trigger GC after cleanup

# API retry configuration
export API_RETRY_COUNT=3               # Number of retry attempts for API calls
export API_RETRY_DELAY=1.0             # Base delay between retries in seconds (exponential backoff)

# Slack Notification
export SLACK_ENABLED=false
export SLACK_BOT_TOKEN='XXXXXXXX'
export SLACK_CHANNEL_ID='XXXXXXXX'

# Feature flags
export DISABLE_PROD_DELETION=true
export DISABLE_NO_LABEL_DELETION=true

# Validation
if [ -z "$HARBOR_USERNAME" ] || [ -z "$HARBOR_PASSWORD" ]; then
    echo "❌ ERROR: Harbor credentials not set!"
    return 1
fi

if ! command -v python3 &> /dev/null; then
    echo "❌ ERROR: python3 is not installed!"
    return 1
fi

# Display configuration
echo ""
echo "✅ Harbor cleanup environment variables loaded:"
echo "🔗 Connection:"
echo "   - Username: $HARBOR_USERNAME"
echo "   - Password: $(echo $HARBOR_PASSWORD | sed 's/./*/g')"  # Mask password
echo ""
echo "📋 Processing mode:"
if [ "$USE_REPO_LIST" = "true" ]; then
    echo "   - Selected repositories: $REPO_LIST"
else
    echo "   - All repositories in project"
fi
echo ""
echo "⚙️  Processing settings:"
echo "   - Parallel processing: $ENABLE_PARALLEL (repo workers: $MAX_WORKERS, page workers: $MAX_PAGE_WORKERS, artifact workers: $((MAX_WORKERS < MAX_ARTIFACT_WORKERS ? MAX_WORKERS : MAX_ARTIFACT_WORKERS)))"
echo "   - Page threshold: $PARALLEL_PAGE_THRESHOLD artifacts"
echo "   - Min artifacts threshold: $MIN_ARTIFACTS_THRESHOLD artifacts"
echo "   - API retry: $API_RETRY_COUNT attempts (delay: ${API_RETRY_DELAY}s)"
echo "   - Log level: $LOG_LEVEL"
echo "   - Verbose artifacts: $VERBOSE_ARTIFACTS"
echo "   - Dry run: $DRY_RUN"
echo ""
echo "🗂️  Retention policies:"
echo "   - Prod artifacts: keep last $RETAIN_LAST_PROD (+ ${PROD_LABEL_CLEANUP_DAYS}d age policy)"
echo "   - Non-prod artifacts: keep last $RETAIN_LAST_NON_PROD (+ ${NON_PROD_LABEL_CLEANUP_DAYS}d age policy)"
echo "   - No-label artifacts: keep last $RETAIN_LAST_NO_LABEL (+ ${EMPTY_LABEL_CLEANUP_DAYS}d age policy)"
echo "   - Other artifacts: keep last $RETAIN_LAST_OTHER (+ ${EMPTY_LABEL_CLEANUP_DAYS}d age policy)"
echo ""
if [ "$DRY_RUN" = "false" ]; then
    echo "⚠️  WARNING: DRY_RUN is disabled - artifacts WILL BE DELETED!"
    echo "   Make sure you've reviewed the dry run output first!"
else
    echo "🛡️  Safe mode: DRY_RUN enabled - no artifacts will be deleted"
fi

echo ""
if [ "$SLACK_ENABLED" = "true" ]; then
    echo "🔔 Slack notifications enabled"
    echo "   - Bot token: $(echo $SLACK_BOT_TOKEN | sed 's/./*/g')"
    echo "   - Channel ID: $SLACK_CHANNEL_ID"
else
    echo "🔕 Slack notifications disabled"
fi

echo ""
echo "🚮 Garbage collection:"
if [ "$TRIGGER_GC_AFTER_CLEANUP" = "true" ]; then
    echo "   ✅ Will trigger GC after cleanup"
else
    echo "   ⭕ GC will not be triggered automatically"
fi
echo ""
echo "🚀 Ready to run: python3 main.py"
echo ""