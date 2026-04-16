"""
Harbor Cleanup Tool - Reporting Module

This module handles all output formatting and summary generation.
"""

import time
from ..core.config import (
    PROJECT, LOG_LEVEL, VERBOSE_ARTIFACTS, DRY_RUN, MAX_WORKERS, MAX_PAGE_WORKERS,
    API_RETRY_COUNT, API_RETRY_DELAY, TRIGGER_GC_AFTER_CLEANUP, ENABLE_PARALLEL, DELETION_SUMMARY_INTERVAL, logger,
    MAX_ARTIFACT_WORKERS
)
from ..core.retention_policy import RetentionPolicy
from .formatting import format_size, format_duration
from ..api.harbor_api import trigger_gc
from .slack_notifier import slack_notifier

def print_startup_banner(script_start_time):
    """Print the startup banner with configuration info"""
    logger.info("🚀 Harbor Cleanup Tool")
    logger.info(f"Project: {PROJECT}")
    logger.info(f"⏱️  Script started at: {time.strftime('%H:%M:%S', time.localtime(script_start_time))}")
    logger.info(f"🔧 Configuration:")
    logger.info(f"  - Log level: {LOG_LEVEL}")
    logger.info(f"  - Verbose artifacts: {VERBOSE_ARTIFACTS}")
    logger.info(f"  - Dry run: {DRY_RUN}")
    
    # Display retention policies using the new system
    retention_policy = RetentionPolicy()
    policy_summary = retention_policy.get_summary()
    for line in policy_summary.split('\n'):
        logger.info(line)
    
    logger.info(f"  - Parallel processing: {'enabled' if ENABLE_PARALLEL else 'disabled'} (repo workers: {MAX_WORKERS}, artifact workers: {min(MAX_WORKERS, MAX_ARTIFACT_WORKERS)}, page workers: {MAX_PAGE_WORKERS})")
    logger.info(f"  - API retry count: {API_RETRY_COUNT} (delay: {API_RETRY_DELAY}s)")
    logger.info(f"  - Deletion progress summaries: every {DELETION_SUMMARY_INTERVAL} deletions")
    logger.info(f"  - Trigger GC after cleanup: {TRIGGER_GC_AFTER_CLEANUP}")
    
    if DRY_RUN:
        logger.info("🛡️  DRY RUN MODE: No artifacts will be actually deleted")
    else:
        logger.warning("⚠️  LIVE MODE: Artifacts WILL BE DELETED!")
    
    logger.info("")

def print_detailed_output(results):
    """Print detailed output from parallel processing"""
    if ENABLE_PARALLEL:
        logger.info(f"{'='*80}")
        logger.info("📝 DETAILED OUTPUT FROM PARALLEL PROCESSING")
        logger.info(f"{'='*80}")
        
        for result in results:
            if result.get('stdout') or result.get('stderr'):
                logger.info(f"-- Output for {result['repo_name']} ---")
                if result.get('stdout'):
                    print(result['stdout'])  # Print raw output without logger formatting
                if result.get('stderr'):
                    logger.error(f"STDERR: {result['stderr']}")

def print_overall_summary(results, script_start_time):
    """Print the overall summary and timing information"""
    logger.info(f"{'='*80}")
    logger.info("🏁 OVERALL SUMMARY")
    logger.info(f"{'='*80}")
    
    total_repos_processed = len(results)
    total_artifacts = sum(r['result']['total_artifacts'] for r in results)
    total_to_delete = sum(r['result']['artifacts_to_delete'] for r in results)
    total_to_keep = sum(r['result']['artifacts_to_keep'] for r in results)
    total_cleanup_space = sum(r['result']['space_to_cleanup'] for r in results)
    total_space = sum(r['result']['total_space'] for r in results)
    
    logger.info(f"Repositories processed: {total_repos_processed:,}")
    logger.info(f"📦 ARTIFACT SUMMARY:")
    logger.info(f"   Total artifacts processed: {total_artifacts:,}")
    logger.info(f"   ├─ Artifacts to delete: {total_to_delete:,} ({(total_to_delete/total_artifacts*100):.1f}%)" if total_artifacts > 0 else f"   ├─ Artifacts to delete: {total_to_delete:,}")
    logger.info(f"   └─ Artifacts to keep: {total_to_keep:,} ({(total_to_keep/total_artifacts*100):.1f}%)" if total_artifacts > 0 else f"   └─ Artifacts to keep: {total_to_keep:,}")
    logger.info(f"💾 STORAGE SUMMARY:")
    logger.info(f"   Total storage: {format_size(total_space)}")
    logger.info(f"   ├─ Space to cleanup: {format_size(total_cleanup_space)}")
    logger.info(f"   └─ Space to remain: {format_size(total_space - total_cleanup_space)}")
    
    if total_cleanup_space > 0 and total_space > 0:
        overall_percentage = (total_cleanup_space / total_space) * 100
        logger.info(f"Overall cleanup percentage: {overall_percentage:.1f}%")
    
    # Aggregate category totals across all repos (prod, non-prod, no-label, other)
    category_totals = {'prod': 0, 'non-prod': 0, 'no-label': 0, 'other': 0}
    for r in results:
        per_repo_totals = r['result'].get('category_totals') or {}
        for cat in category_totals.keys():
            category_totals[cat] += per_repo_totals.get(cat, 0)
    
    logger.info("🏷️ CATEGORY DISTRIBUTION (across all repositories):")
    logger.info(f"   ├─ prod: {category_totals['prod']:,}")
    logger.info(f"   ├─ non-prod: {category_totals['non-prod']:,}")
    logger.info(f"   ├─ no-label: {category_totals['no-label']:,}")
    logger.info(f"   └─ other: {category_totals['other']:,}")
    
    # Calculate processing time statistics
    processing_times = [r['result'].get('processing_duration', 0) for r in results if r['success']]
    if processing_times:
        total_processing_time = sum(processing_times)
        avg_processing_time = total_processing_time / len(processing_times)
        max_processing_time = max(processing_times)
        min_processing_time = min(processing_times)
        
        logger.info("⏱️  PROCESSING TIME STATISTICS:")
        logger.info(f"Total processing time: {format_duration(total_processing_time)}")
        logger.info(f"Average processing time per repo: {format_duration(avg_processing_time)}")
        logger.info(f"Fastest repository: {format_duration(min_processing_time)}")
        logger.info(f"Slowest repository: {format_duration(max_processing_time)}")
    
    if results:
        logger.info("📋 Summary by repository:")
        for result in results:
            if result['result']['total_artifacts'] > 0:
                cleanup_pct = (result['result']['space_to_cleanup'] / result['result']['total_space'] * 100) if result['result']['total_space'] > 0 else 0
                status = "✅" if result['success'] else "❌"
                
                # Format timing information for summary
                processing_time = result['result'].get('processing_duration', 0)
                queue_time = result['result'].get('queue_time', 0)
                
                if queue_time > 1:  # Show queue time if significant
                    timing_info = f", processing: {format_duration(processing_time)}, queue: {format_duration(queue_time)}"
                else:
                    timing_info = f", {format_duration(processing_time)}"
                
                logger.info(f"  {status} {result['repo_name']}: {result['result']['artifacts_to_delete']}/{result['result']['total_artifacts']} artifacts ({format_size(result['result']['space_to_cleanup'])}, {cleanup_pct:.1f}%{timing_info})")
    
    script_end_time = time.time()
    total_elapsed_time = script_end_time - script_start_time
    
    logger.info("⏱️  TIMING SUMMARY:")
    logger.info(f"Script started at: {time.strftime('%H:%M:%S', time.localtime(script_start_time))}")
    logger.info(f"Script ended at: {time.strftime('%H:%M:%S')}")
    logger.info(f"Total execution time: {format_duration(total_elapsed_time)}")
    
    if DRY_RUN:
        logger.info("⚠️  This was a DRY RUN - no artifacts were actually deleted!")
    else:
        logger.info("🗑️  LIVE RUN completed - artifacts were actually deleted!")
    logger.info(f"{'='*80}")

def handle_post_cleanup_gc(results):
    """Handle garbage collection after cleanup if enabled and return GC status"""
    if not TRIGGER_GC_AFTER_CLEANUP:
        return None
        
    logger.info(f"{'='*60}")
    logger.info("🚮 POST-CLEANUP GARBAGE COLLECTION")
    logger.info(f"{'='*60}")
    
    total_cleanup_space = sum(r['result']['space_to_cleanup'] for r in results if r['success'])

    # Safety check for very large cleanups
    cleanup_size_gb = total_cleanup_space / (1024**3)  # Convert to GB
    if cleanup_size_gb > 50000:  # > 50TB
        logger.warning(f"⚠️  LARGE CLEANUP DETECTED: {format_size(total_cleanup_space)} ({cleanup_size_gb:.1f} GB)")
        logger.warning("⚠️  This is a very large cleanup that might stress Harbor's GC system")
        logger.warning("⚠️  Consider:")
        logger.warning("   - Running cleanup in smaller batches")
        logger.warning("   - Monitoring Harbor UI for GC progress")
        logger.warning("   - Running GC during off-peak hours")
        logger.warning("   - Checking Harbor's GC timeout settings")

        if not DRY_RUN:
            logger.warning("⚠️  Proceeding with GC trigger anyway...")

    gc_success = trigger_gc()

    if gc_success:
        logger.info("✅ Garbage collection triggered successfully")
        if not DRY_RUN:
            logger.info("ℹ️  Note: GC runs asynchronously. Check Harbor UI for progress.")
            if cleanup_size_gb > 10000:  # > 10TB
                logger.info("ℹ️  Large cleanup detected - GC may take several hours to complete")
                logger.info("ℹ️  Monitor Harbor UI and system resources during GC")
    else:
        logger.warning("⚠️  Failed to trigger garbage collection")

    # Return GC status for Slack notification
    return {'success': gc_success, 'cleanup_size_gb': cleanup_size_gb}