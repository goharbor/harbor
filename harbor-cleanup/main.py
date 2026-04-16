#!/usr/bin/env python3
"""
Harbor Cleanup Tool - Main Entry Point

A comprehensive Python tool for cleaning up Harbor registry artifacts 
based on configurable retention policies.

Features modular architecture for maintainability and extensibility.
"""

import time
import sys

# Import all modules
from harbor_cleanup.core.config import logger
from harbor_cleanup.utils.formatting import setup_signal_handlers, shutdown_requested
from harbor_cleanup.core.processor import (
    get_selected_repositories, 
    process_selected_repositories, 
    process_all_repositories,
    reset_progress_tracker,
    get_progress_summary
)
from harbor_cleanup.utils.reporting import (
    print_startup_banner,
    print_detailed_output,
    print_overall_summary,
    handle_post_cleanup_gc
)
from harbor_cleanup.utils.slack_notifier import slack_notifier

def main():
    """Main orchestration function"""
    script_start_time = time.time()
    
    # Set up signal handlers for graceful shutdown
    setup_signal_handlers()
    
    # Reset progress tracker for this cleanup session
    reset_progress_tracker()
    
    # Print startup information
    print_startup_banner(script_start_time)
    
    try:
    # Determine processing mode and execute
        selected_repos = get_selected_repositories()
        
        # Send Slack notification that cleanup is starting
        slack_notifier.send_cleanup_start(selected_repos)
    
        if selected_repos:
            # Process selected repositories
            results = process_selected_repositories(selected_repos)
        else:
            # Process all repositories
            results = process_all_repositories()
        
        # Print detailed output and summary
        print_detailed_output(results)
        
        # Check if we were interrupted
        if shutdown_requested:
            logger.warning(f"🚫 INTERRUPTED: Script was stopped by user request")
            logger.warning(f"✅ Completed processing: {len([r for r in results if r['success']])} repositories")
            logger.warning(f"❌ Cancelled/Failed: {len([r for r in results if not r['success']])} repositories")
            logger.warning("📊 Partial results shown below:")
        
        print_overall_summary(results, script_start_time)
        
        # Print final deletion progress summary
        final_progress = get_progress_summary()
        if final_progress['total_deletions'] > 0:
            logger.info(f"🎯 FINAL DELETION SUMMARY:")
            logger.info(f"   ├─ Total deletion attempts: {final_progress['total_deletions']:,}")
            logger.info(f"   ├─ Successful deletions: {final_progress['successful_deletions']:,}")
            logger.info(f"   ├─ Failed deletions: {final_progress['failed_deletions']:,}")
            logger.info(f"   ├─ Overall success rate: {final_progress['success_rate']:.1f}%")
            logger.info(f"   ├─ Total space freed: {final_progress['total_space_freed_formatted']}")
            logger.info(f"   ├─ Processing time: {final_progress['elapsed_time_formatted']}")
            logger.info(f"   └─ Average deletion rate: {final_progress['avg_rate']:.1f} deletions/sec")
    
    # Optionally trigger garbage collection after cleanup
        gc_status = handle_post_cleanup_gc(results)
        
        # Send Slack notification that cleanup completed successfully
        slack_notifier.send_cleanup_complete(results, script_start_time, interrupted=shutdown_requested)
        
        # Send GC notification after completion notification (if GC was triggered)
        if gc_status:
            slack_notifier.send_gc_notification(gc_status['success'], gc_status['cleanup_size_gb'])
        
    except KeyboardInterrupt:
        logger.error("🚫 Script interrupted by user")
        # Send Slack notification about interruption
        slack_notifier.send_error_notification("Script interrupted by user (Ctrl+C)")
        sys.exit(1)
    except Exception as e:
        logger.error(f"❌ Unexpected error: {e}")
        # Send Slack notification about the error
        slack_notifier.send_error_notification(str(e))
        sys.exit(1)

if __name__ == "__main__":
    main()