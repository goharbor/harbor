"""
Harbor Cleanup Tool - Configuration Module

This module centralizes all configuration settings and environment variable handling.
"""

import os
import sys
import logging

# Setup logging first
LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO").upper()
VERBOSE_ARTIFACTS = os.getenv("VERBOSE_ARTIFACTS", "true").lower() in ("true", "1", "yes", "on")
DRY_RUN = os.getenv("DRY_RUN", "true").lower() in ("true", "1", "yes", "on")

# Convert log level string to logging constant
log_level_map = {
    'DEBUG': logging.DEBUG,
    'INFO': logging.INFO,
    'WARNING': logging.WARNING,
    'ERROR': logging.ERROR
}
log_level = log_level_map.get(LOG_LEVEL, logging.INFO)

# Enhanced logging format with better structure
logging.basicConfig(
    level=log_level,
    format='%(asctime)s - %(levelname)s - %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S',
    handlers=[
        logging.StreamHandler(sys.stdout)
    ]
)

# Create custom logger with module-specific formatting
logger = logging.getLogger('harbor-cleanup')

# Add performance logging flag
PERFORMANCE_LOGGING = os.getenv("PERFORMANCE_LOGGING", "false").lower() in ("true", "1", "yes", "on")

# Deletion progress summary interval
DELETION_SUMMARY_INTERVAL = int(os.getenv("DELETION_SUMMARY_INTERVAL", "100"))

# Harbor connection settings
HARBOR_URL = os.getenv("HARBOR_URL", "https://harbor.razorpay.com")
PROJECT = os.getenv("HARBOR_PROJECT", "razorpay")
USERNAME = os.getenv("HARBOR_USERNAME")
PASSWORD = os.getenv("HARBOR_PASSWORD")
VERIFY_SSL = False  # Set to True in prod

# Validate required credentials
if not USERNAME:
    logger.error("❌ HARBOR_USERNAME environment variable is required")
    sys.exit(1)
if not PASSWORD:
    logger.error("❌ HARBOR_PASSWORD environment variable is required")
    sys.exit(1)

# Repository selection
REPO_LIST = os.getenv("REPO_LIST", "relay,ledger,authz")
USE_REPO_LIST = os.getenv("USE_REPO_LIST", "true").lower() in ("true", "1", "yes", "on")

# Parallel processing configuration
MAX_WORKERS = int(os.getenv("MAX_WORKERS", "5"))  # Number of parallel repository processes
ENABLE_PARALLEL = os.getenv("ENABLE_PARALLEL", "true").lower() in ("true", "1", "yes", "on")
# Page-level parallelism configuration
MAX_PAGE_WORKERS = int(os.getenv("MAX_PAGE_WORKERS", "3"))  # Number of parallel page fetches per repository
PARALLEL_PAGE_THRESHOLD = int(os.getenv("PARALLEL_PAGE_THRESHOLD", "500"))
# Artifact-level parallelism configuration
MAX_ARTIFACT_WORKERS = int(os.getenv("MAX_ARTIFACT_WORKERS", "5"))  # Max concurrent artifact deletions per repo

# Repository filtering
MIN_ARTIFACTS_THRESHOLD = int(os.getenv("MIN_ARTIFACTS_THRESHOLD", "100"))

# Garbage collection configuration
TRIGGER_GC_AFTER_CLEANUP = os.getenv("TRIGGER_GC_AFTER_CLEANUP", "false").lower() in ("true", "1", "yes", "on")
GC_DELETE_UNTAGGED = os.getenv("GC_DELETE_UNTAGGED", "false").lower() in ("true", "1", "yes", "on")
GC_WORKERS = int(os.getenv("GC_WORKERS", "5"))

# API retry configuration
API_RETRY_COUNT = int(os.getenv("API_RETRY_COUNT", "3"))  # Number of retry attempts for API calls
API_RETRY_DELAY = float(os.getenv("API_RETRY_DELAY", "1.0"))  # Delay between retries in seconds 

# Slack notification configuration
SLACK_ENABLED = os.getenv("SLACK_ENABLED", "false").lower() in ("true", "1", "yes", "on")
SLACK_BOT_TOKEN = os.getenv("SLACK_BOT_TOKEN")  # Slack Bot User OAuth Token (xoxb-...)
SLACK_CHANNEL_ID = os.getenv("SLACK_CHANNEL_ID")  # Slack Channel ID (C1234567890 or #channel-name)
HARBOR_ENV = os.getenv("HARBOR_ENV", "NotDefined")

# Feature flags
DISABLE_PROD_DELETION = os.getenv("DISABLE_PROD_DELETION", "false").lower() in ("true", "1", "yes", "on")
DISABLE_NO_LABEL_DELETION = os.getenv("DISABLE_NO_LABEL_DELETION", "false").lower() in ("true", "1", "yes", "on") 