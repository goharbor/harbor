#!/usr/bin/env python3
"""
Harbor Cleanup Tool - CLI Interface

This module provides a comprehensive command-line interface with subcommands
for different cleanup operations and management tasks.
"""

import argparse
import sys
import os
from typing import List, Optional

from .core.config import logger
from .utils.checkpoint_manager import checkpoint_manager
from .utils.metrics_collector import metrics_collector

def create_parser() -> argparse.ArgumentParser:
    """Create and configure the argument parser"""
    parser = argparse.ArgumentParser(
        prog='harbor-cleanup',
        description='Harbor Registry Cleanup Tool - Automated artifact retention management',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Dry run cleanup for specific repositories
  harbor-cleanup cleanup --repos authz,api,ledger --dry-run
  
  # Live cleanup with custom retention
  harbor-cleanup cleanup --live --keep-count 10 --keep-days 30
  
  # Resume interrupted cleanup
  harbor-cleanup resume --auto
  
  # Show cleanup status and metrics
  harbor-cleanup status --detailed
  
  # Cleanup specific artifact
  harbor-cleanup delete-artifact authz sha256:abc123...
  
  # Test retention policies
  harbor-cleanup test-policy --repos authz --verbose

Environment Variables:
  HARBOR_USERNAME, HARBOR_PASSWORD, HARBOR_PROJECT - Required credentials
  HARBOR_URL - Harbor instance URL (default: https://harbor.razorpay.com)
  MAX_WORKERS, ENABLE_PARALLEL - Performance tuning
  SLACK_ENABLED, SLACK_BOT_TOKEN - Optional notifications
        """
    )
    
    # Global options
    parser.add_argument('--version', action='version', version='%(prog)s 2.0.0')
    parser.add_argument('--config-file', help='Path to configuration file')
    parser.add_argument('--log-level', choices=['DEBUG', 'INFO', 'WARNING', 'ERROR'],
                       help='Set logging level')
    
    # Create subparsers
    subparsers = parser.add_subparsers(dest='command', help='Available commands')
    
    # Cleanup command
    cleanup_parser = subparsers.add_parser('cleanup', help='Run artifact cleanup')
    cleanup_parser.add_argument('--repos', help='Comma-separated list of repositories')
    cleanup_parser.add_argument('--dry-run', action='store_true', help='Run in dry-run mode (default)')
    cleanup_parser.add_argument('--live', action='store_true', help='Run in live mode (actually delete)')
    cleanup_parser.add_argument('--keep-count', type=int, help='Number of artifacts to keep per category')
    cleanup_parser.add_argument('--keep-days', type=int, help='Keep artifacts newer than N days')
    cleanup_parser.add_argument('--parallel', action='store_true', help='Enable parallel processing')
    cleanup_parser.add_argument('--workers', type=int, help='Number of worker threads')
    cleanup_parser.add_argument('--force-gc', action='store_true', help='Force garbage collection after cleanup')
    cleanup_parser.add_argument('--checkpoint', action='store_true', help='Enable checkpointing')
    
    # Resume command
    resume_parser = subparsers.add_parser('resume', help='Resume interrupted cleanup')
    resume_parser.add_argument('--auto', action='store_true', help='Auto-resume from checkpoint')
    resume_parser.add_argument('--checkpoint-file', help='Path to checkpoint file')
    resume_parser.add_argument('--list', action='store_true', help='List available checkpoints')
    
    # Status command
    status_parser = subparsers.add_parser('status', help='Show cleanup status and metrics')
    status_parser.add_argument('--detailed', action='store_true', help='Show detailed metrics')
    status_parser.add_argument('--repos', help='Check specific repositories')
    status_parser.add_argument('--json', action='store_true', help='Output in JSON format')
    
    # Delete specific artifact
    delete_parser = subparsers.add_parser('delete-artifact', help='Delete specific artifact by digest')
    delete_parser.add_argument('repository', help='Repository name')
    delete_parser.add_argument('digest', help='Artifact digest (sha256:...)')
    delete_parser.add_argument('--force', action='store_true', help='Force deletion')
    delete_parser.add_argument('--dry-run', action='store_true', help='Simulate deletion')
    
    # Test retention policies
    test_parser = subparsers.add_parser('test-policy', help='Test retention policies')
    test_parser.add_argument('--repos', help='Repositories to test')
    test_parser.add_argument('--verbose', action='store_true', help='Verbose output')
    test_parser.add_argument('--export', help='Export policy test results to file')
    
    # List repositories
    list_parser = subparsers.add_parser('list-repos', help='List repositories')
    list_parser.add_argument('--filter', help='Filter repositories by name pattern')
    list_parser.add_argument('--sort', choices=['name', 'size', 'artifacts'], default='name')
    list_parser.add_argument('--limit', type=int, help='Limit number of results')
    
    # Metrics command
    metrics_parser = subparsers.add_parser('metrics', help='Show cleanup metrics')
    metrics_parser.add_argument('--session-id', help='Show metrics for specific session')
    metrics_parser.add_argument('--export', help='Export metrics to file')
    metrics_parser.add_argument('--last', type=int, help='Show last N sessions')
    
    # Configuration management
    config_parser = subparsers.add_parser('config', help='Configuration management')
    config_subparsers = config_parser.add_subparsers(dest='config_action')
    
    config_subparsers.add_parser('show', help='Show current configuration')
    config_subparsers.add_parser('validate', help='Validate configuration')
    config_subparsers.add_parser('test-connection', help='Test Harbor connection')
    
    return parser

def handle_cleanup_command(args) -> int:
    """Handle cleanup command"""
    from .main import main as cleanup_main
    
    # Override environment variables with CLI arguments
    if args.repos:
        os.environ['REPO_LIST'] = args.repos
        os.environ['USE_REPO_LIST'] = 'true'
    
    if args.live:
        os.environ['DRY_RUN'] = 'false'
    elif args.dry_run:
        os.environ['DRY_RUN'] = 'true'
    
    if args.keep_count:
        # Set retention policy overrides
        for category in ['prod', 'non-prod', 'no-label', 'other']:
            os.environ[f'RETENTION_COUNT_{category.upper().replace("-", "_")}'] = str(args.keep_count)
    
    if args.keep_days:
        for category in ['prod', 'non-prod', 'no-label', 'other']:
            os.environ[f'RETENTION_DAYS_{category.upper().replace("-", "_")}'] = str(args.keep_days)
    
    if args.parallel:
        os.environ['ENABLE_PARALLEL'] = 'true'
    
    if args.workers:
        os.environ['MAX_WORKERS'] = str(args.workers)
    
    if args.force_gc:
        os.environ['TRIGGER_GC_AFTER_CLEANUP'] = 'true'
    
    if args.checkpoint:
        os.environ['CHECKPOINT_ENABLED'] = 'true'
    
    # Run cleanup
    try:
        cleanup_main()
        return 0
    except KeyboardInterrupt:
        logger.error("Cleanup interrupted by user")
        return 1
    except Exception as e:
        logger.error(f"Cleanup failed: {e}")
        return 1

def handle_resume_command(args) -> int:
    """Handle resume command"""
    if args.list:
        # List available checkpoints
        resume_info = checkpoint_manager.get_resume_info()
        if resume_info:
            logger.info("📊 Available checkpoint:")
            logger.info(f"   ├─ Session ID: {resume_info['session_id']}")
            logger.info(f"   ├─ Progress: {resume_info['completion_percentage']:.1f}%")
            logger.info(f"   ├─ Processed: {resume_info['processed_repositories']}")
            logger.info(f"   ├─ Remaining: {resume_info['remaining_repositories']}")
            logger.info(f"   └─ File: {resume_info['checkpoint_file']}")
        else:
            logger.info("No checkpoint files found")
        return 0
    
    if args.auto:
        os.environ['AUTO_RESUME'] = 'true'
    
    if args.checkpoint_file:
        os.environ['CHECKPOINT_FILE'] = args.checkpoint_file
    
    # Run cleanup with resume enabled
    return handle_cleanup_command(args)

def handle_status_command(args) -> int:
    """Handle status command"""
    from .api.harbor_api import get_cached_repositories, get_repository_info
    from .utils.formatting import format_size
    
    try:
        if args.repos:
            repos = args.repos.split(',')
        else:
            repos = get_cached_repositories()[:10]  # Limit to first 10 for status
        
        logger.info("📊 Harbor Cleanup Status")
        logger.info("=" * 50)
        
        total_artifacts = 0
        total_size = 0
        
        for repo in repos:
            try:
                repo_info = get_repository_info(repo)
                if repo_info:
                    total_artifacts += repo_info.get('artifact_count', 0)
                    # Note: size info would need to be calculated from artifacts
                    logger.info(f"✅ {repo}: {repo_info.get('artifact_count', 0):,} artifacts")
                else:
                    logger.warning(f"⚠️  {repo}: Not found or no access")
            except Exception as e:
                logger.error(f"❌ {repo}: Error - {e}")
        
        logger.info("=" * 50)
        logger.info(f"Total artifacts in checked repositories: {total_artifacts:,}")
        
        # Show checkpoint status
        resume_info = checkpoint_manager.get_resume_info()
        if resume_info:
            logger.info("\n🔄 Checkpoint Status:")
            logger.info(f"   Resumable session found: {resume_info['completion_percentage']:.1f}% complete")
        
        return 0
        
    except Exception as e:
        logger.error(f"Status check failed: {e}")
        return 1

def handle_delete_artifact_command(args) -> int:
    """Handle delete-artifact command (digest-based deletion only)"""
    from .api.harbor_api import delete_artifact
    from .core.config import PROJECT
    
    # Validate that the input is a digest (starts with sha256:)
    if not args.digest.startswith('sha256:'):
        logger.error(f"❌ Invalid digest format: {args.digest}")
        logger.error(f"   💡 Digest must start with 'sha256:' (e.g., sha256:abc123...)")
        logger.error(f"   💡 Tag-based deletion is no longer supported - use digest only")
        return 1
    
    if args.dry_run:
        logger.info(f"🔍 [DRY RUN] Would delete artifact: {args.repository}/{args.digest[:12]}...")
        return 0
    
    try:
        success = delete_artifact(PROJECT, args.repository, args.digest)
        if success:
            logger.info(f"✅ Successfully deleted artifact: {args.repository}/{args.digest[:12]}...")
            return 0
        else:
            logger.error(f"❌ Failed to delete artifact: {args.repository}/{args.digest[:12]}...")
            return 1
    except Exception as e:
        logger.error(f"Delete failed: {e}")
        return 1

def handle_test_policy_command(args) -> int:
    """Handle test-policy command"""
    from .core.retention_policy import RetentionPolicy
    from .api.harbor_api import get_cached_repositories, list_artifacts
    
    try:
        policy = RetentionPolicy()
        
        if args.repos:
            repos = args.repos.split(',')
        else:
            repos = get_cached_repositories()[:5]  # Test first 5 repos
        
        logger.info("🧪 Testing Retention Policies")
        logger.info("=" * 50)
        
        for repo in repos:
            logger.info(f"\n📋 Testing repository: {repo}")
            
            try:
                artifacts = list_artifacts("razorpay", repo)
                if not artifacts:
                    logger.info("   No artifacts found")
                    continue
                
                # Test policy on first few artifacts
                test_artifacts = artifacts[:10] if not args.verbose else artifacts
                
                categories = {'prod': 0, 'non-prod': 0, 'no-label': 0, 'other': 0}
                
                for artifact in test_artifacts:
                    category = policy.categorize_artifact(artifact)
                    categories[category] += 1
                    
                    if args.verbose:
                        logger.info(f"   📦 {artifact.get('digest', 'unknown')[:12]}... → {category}")
                
                logger.info(f"   📊 Category distribution: {categories}")
                
            except Exception as e:
                logger.error(f"   ❌ Error testing {repo}: {e}")
        
        return 0
        
    except Exception as e:
        logger.error(f"Policy test failed: {e}")
        return 1

def handle_config_command(args) -> int:
    """Handle config command"""
    from .core.config import (
        HARBOR_URL, PROJECT, USERNAME, ENABLE_PARALLEL, MAX_WORKERS,
        DRY_RUN, VERBOSE_ARTIFACTS
    )
    
    if args.config_action == 'show':
        logger.info("⚙️  Current Configuration:")
        logger.info(f"   ├─ Harbor URL: {HARBOR_URL}")
        logger.info(f"   ├─ Project: {PROJECT}")
        logger.info(f"   ├─ Username: {USERNAME}")
        logger.info(f"   ├─ Parallel processing: {ENABLE_PARALLEL}")
        logger.info(f"   ├─ Max workers: {MAX_WORKERS}")
        logger.info(f"   ├─ Dry run: {DRY_RUN}")
        logger.info(f"   └─ Verbose artifacts: {VERBOSE_ARTIFACTS}")
        
    elif args.config_action == 'test-connection':
        from .api.harbor_api import get_cached_repositories
        try:
            repos = get_cached_repositories()
            logger.info(f"✅ Connection successful: Found {len(repos)} repositories")
        except Exception as e:
            logger.error(f"❌ Connection failed: {e}")
            return 1
    
    return 0

def main() -> int:
    """Main CLI entry point"""
    parser = create_parser()
    args = parser.parse_args()
    
    # Override log level if specified
    if args.log_level:
        os.environ['LOG_LEVEL'] = args.log_level
    
    # Handle no command
    if not args.command:
        parser.print_help()
        return 1
    
    # Route to appropriate handler
    if args.command == 'cleanup':
        return handle_cleanup_command(args)
    elif args.command == 'resume':
        return handle_resume_command(args)
    elif args.command == 'status':
        return handle_status_command(args)
    elif args.command == 'delete-artifact':
        return handle_delete_artifact_command(args)
    elif args.command == 'test-policy':
        return handle_test_policy_command(args)
    elif args.command == 'config':
        return handle_config_command(args)
    else:
        logger.error(f"Unknown command: {args.command}")
        return 1

if __name__ == '__main__':
    sys.exit(main()) 