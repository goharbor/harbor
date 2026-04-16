"""
Harbor Cleanup Tool - Processor Module

This module handles repository processing logic and parallel execution.
"""

import time
import threading
import io
from datetime import datetime
from contextlib import redirect_stdout, redirect_stderr
from concurrent.futures import ThreadPoolExecutor, as_completed

from .config import (
    PROJECT, MAX_WORKERS, ENABLE_PARALLEL, MIN_ARTIFACTS_THRESHOLD,
    VERBOSE_ARTIFACTS, DRY_RUN, REPO_LIST, USE_REPO_LIST, DELETION_SUMMARY_INTERVAL, logger,
    DISABLE_PROD_DELETION, DISABLE_NO_LABEL_DELETION, MAX_ARTIFACT_WORKERS
)
from ..api.harbor_api import (
    get_cached_repositories, get_repository_info, list_artifacts, delete_artifact_with_fallback
)
from .retention_policy import RetentionPolicy
from ..utils.formatting import format_size, format_duration, shutdown_requested
from ..utils.logging_utils import DeletionProgressTracker

# Configuration for artifact-level parallelism
ARTIFACT_WORKERS = min(MAX_WORKERS, MAX_ARTIFACT_WORKERS)

# Initialize global retention policy
RETENTION_POLICY = RetentionPolicy()

# Global progress tracker for deletion summaries
global_progress_tracker = DeletionProgressTracker(summary_interval=DELETION_SUMMARY_INTERVAL)

def reset_progress_tracker():
    """Reset the global progress tracker for a new cleanup session"""
    global global_progress_tracker
    global_progress_tracker = DeletionProgressTracker(summary_interval=DELETION_SUMMARY_INTERVAL)

def get_progress_summary():
    """Get the final progress summary"""
    return global_progress_tracker.get_final_summary()

def delete_artifact_parallel(project, repo, artifact_data):
    """Delete a single artifact with proper error handling for parallel execution"""
    try:
        artifact = artifact_data['artifact']
        digest = artifact_data['digest']
        tags = artifact_data['tags']

        # Actually delete the artifact if not in dry run mode
        logger.debug(f"🎯 Parallel deletion: digest-based deletion only")
        deletion_success = delete_artifact_with_fallback(project, repo, digest, tags, artifact)

        # Record progress for the global tracker
        global_progress_tracker.record_deletion(
            success=deletion_success or DRY_RUN,
            space_freed=artifact_data['size'] if (deletion_success or DRY_RUN) else 0
        )

        return {
            'artifact_data': artifact_data,
            'success': deletion_success or DRY_RUN,
            'actual_deletion': deletion_success
        }
    except Exception as e:
        logger.error(f"❌ Exception during parallel deletion of {digest[:12]}: {e}")
        
        # Record failed deletion
        global_progress_tracker.record_deletion(success=False, space_freed=0)
        
        return {
            'artifact_data': artifact_data,
            'success': False,
            'actual_deletion': False,
            'error': str(e)
        }

def process_repository(project, repo_name, artifact_count=None):
    """Process cleanup for a single repository"""
    start_time = time.time()

    # Strip project prefix if present, preserving nested paths (e.g. "razorpay/cache/infra-tools" -> "cache/infra-tools")
    prefix = f"{project}/"
    repo = repo_name[len(prefix):] if repo_name.startswith(prefix) else repo_name
    
    logger.info(f"{'='*80}")
    logger.info(f"🗂️  PROCESSING REPOSITORY: {project}/{repo}")
    if artifact_count is not None:
        logger.info(f"📊 Expected artifacts: {artifact_count}")
    logger.info(f"⏱️  Started at: {time.strftime('%H:%M:%S')}")
    logger.info(f"{'='*80}")
    
    artifacts = list_artifacts(project, repo, artifact_count)
    if not artifacts:
        logger.warning(f"⚠️  No artifacts found in {project}/{repo}, skipping...")
        return {
            'repo': repo,
            'total_artifacts': 0,
            'artifacts_to_delete': 0,
            'artifacts_to_keep': 0,
            'space_to_cleanup': 0,
            'total_space': 0
        }
    
    logger.info(f"Found {len(artifacts)} artifacts in repository {project}/{repo}")

    counter = 0
    total_space_to_cleanup = 0
    artifacts_to_delete = []
    artifacts_to_keep = []
    
    # Category tracking
    category_stats = {
        'prod': {'total': 0, 'to_delete': 0, 'total_size': 0, 'delete_size': 0, 'artifacts': []},
        'non-prod': {'total': 0, 'to_delete': 0, 'total_size': 0, 'delete_size': 0, 'artifacts': []},
        'no-label': {'total': 0, 'to_delete': 0, 'total_size': 0, 'delete_size': 0, 'artifacts': []},
        'other': {'total': 0, 'to_delete': 0, 'total_size': 0, 'delete_size': 0, 'artifacts': []}
    }
    
    # Category retention mapping
    category_retention = {
        'prod': RETENTION_POLICY.get_count_policy('prod'),
        'non-prod': RETENTION_POLICY.get_count_policy('non-prod'),
        'no-label': RETENTION_POLICY.get_count_policy('no-label'),
        'other': RETENTION_POLICY.get_count_policy('other')
    }
    
    # First pass: categorize all artifacts and collect them
    for artifact in artifacts:
        counter += 1
        
        # Get artifact details that we always need
        digest = artifact.get("digest", "unknown")
        size_bytes = artifact.get("size", 0)
        tags = artifact.get("tags", [])
        labels = artifact.get("labels", [])
        
        if VERBOSE_ARTIFACTS:
            logger.debug(f"--- Processing artifact {counter} of {len(artifacts)} ---")
            
            # Print artifact digest for identification
            logger.debug(f"Digest: {digest}")
            
            # Print artifact size
            size_formatted = format_size(size_bytes)
            logger.debug(f"Size: {size_formatted}")
            
            # Process tags
            if tags:
                logger.debug("Tags:")
                for tag in tags:
                    tag_name = tag.get("name", "unknown")
                    logger.debug(f"  - {tag_name}")
            else:
                logger.debug("Tags: none")
            
            # Process labels and categorize
            if labels:
                logger.debug("Labels:")
                for label in labels:
                    label_name = label.get("name", "unknown")
                    logger.debug(f"  - {label_name}")
            else:
                logger.debug("Labels: not-set")
        elif counter % 100 == 0:
            # Show progress every 100 artifacts when not in verbose mode
            logger.info(f"📊 Processed {counter} of {len(artifacts)} artifacts in {project}/{repo}...")
        
        # Determine category
        category = RETENTION_POLICY.categorize_artifact(labels)
        
        # Add artifact to category for sorting
        artifact_data = {
            'artifact': artifact,
            'digest': digest,
            'size': size_bytes,
            'tags': [tag.get("name", "unknown") for tag in tags] if tags else [],
            'category': category
        }
        
        category_stats[category]['artifacts'].append(artifact_data)
        category_stats[category]['total'] += 1
        category_stats[category]['total_size'] += size_bytes
    
    # Second pass: Apply retention policies per category and collect all artifacts for deletion
    logger.info(f"{'='*60}")
    logger.info(f"📋 APPLYING RETENTION POLICIES")
    logger.info(f"{'='*60}")
    
    # Collect all artifacts marked for deletion across all categories
    all_artifacts_marked_for_deletion = []
    
    for category, stats in category_stats.items():
        if stats['total'] == 0:
            continue
            
        logger.info(f"🏷️  Processing {category} category ({stats['total']} artifacts)")
        
        # Sort artifacts by timestamp (most recent first)
        # Use pull_time primarily, fall back to push_time
        def get_sort_time(artifact_data):
            artifact = artifact_data['artifact']
            pull_time_str = artifact.get("pull_time")
            push_time_str = artifact.get("push_time")
            
            time_to_use = None
            # Priority: pull_time first, then push_time
            if pull_time_str and pull_time_str != "0001-01-01T00:00:00.000Z":
                time_to_use = pull_time_str
            elif push_time_str and push_time_str != "0001-01-01T00:00:00.000Z":
                time_to_use = push_time_str
            
            if time_to_use:
                try:
                    time_to_use = time_to_use.split('.')[0] + 'Z'
                    return datetime.fromisoformat(time_to_use.replace('Z', '+00:00'))
                except:
                    pass
            
            # Fallback to epoch if no valid time
            return datetime.fromtimestamp(0, tz=datetime.now().astimezone().tzinfo)
        
        sorted_artifacts = sorted(stats['artifacts'], key=get_sort_time, reverse=True)
        
        # Step 1: Evaluate retention policies and mark artifacts for deletion
        category_artifacts_marked_for_deletion = []
        
        for i, artifact_data in enumerate(sorted_artifacts):
            artifact = artifact_data['artifact']
            digest = artifact_data['digest']
            tags = artifact_data['tags']
            
            # Format artifact info for display
            digest_short = digest[:12] if digest != "unknown" else "unknown"
            tags_str = ", ".join(tags) if tags else "none"
            label_info = f"{category} label" if category != "no-label" else "no label"
            
            # Get timestamps for display
            pull_time_str = artifact.get("pull_time", "never")
            push_time_str = artifact.get("push_time", "unknown")
            
            artifact_info = f"🔍 Artifact {digest_short}... | Tags: {tags_str} | Label: {label_info} | Last Pull: {pull_time_str} | Push: {push_time_str}"
            
            # Policy 1: Always retain the last N artifacts per category
            if i < category_retention[category]:
                logger.info(f"✅ KEEP (position {i+1} of last {category_retention[category]}) | {artifact_info}")
                artifacts_to_keep.append(artifact_data)
                continue
            
            # If prod deletion is disabled, keep all prod artifacts beyond count policy too
            if category == 'prod' and DISABLE_PROD_DELETION:
                logger.info(f"🛑 KEEP (prod deletions disabled) | {artifact_info}")
                artifacts_to_keep.append(artifact_data)
                continue
            
            # If no-label deletion is disabled, keep all no-label artifacts beyond count policy too
            if category == 'no-label' and DISABLE_NO_LABEL_DELETION:
                logger.info(f"🛑 KEEP (no-label deletions disabled) | {artifact_info}")
                artifacts_to_keep.append(artifact_data)
                continue
            
            # Policy 2: Apply age-based retention for older artifacts
            should_delete = RETENTION_POLICY.should_delete_by_age(artifact, category)
            
            if should_delete:
                logger.info(f"🗑️  DELETE (age policy) | {artifact_info}")
                category_artifacts_marked_for_deletion.append(artifact_data)
            else:
                logger.info(f"✅ KEEP (age policy) | {artifact_info}")
                artifacts_to_keep.append(artifact_data)
        
        # Add category artifacts to the global list
        all_artifacts_marked_for_deletion.extend(category_artifacts_marked_for_deletion)
        
        # Update category stats for reporting
        stats['to_delete'] = len(category_artifacts_marked_for_deletion)
        stats['delete_size'] = sum(artifact['size'] for artifact in category_artifacts_marked_for_deletion)
    
    # Step 2: Sort all artifacts marked for deletion by size in reverse order (largest first)
    # This ensures the largest artifacts across all categories are deleted first for maximum space cleanup
    if all_artifacts_marked_for_deletion:
        all_artifacts_marked_for_deletion.sort(key=lambda x: x['size'], reverse=True)
        total_size_to_delete = sum(artifact['size'] for artifact in all_artifacts_marked_for_deletion)
        logger.info(f"{'='*60}")
        logger.info(f"📦 CROSS-CATEGORY SIZE-BASED SORTING")
        logger.info(f"{'='*60}")
        logger.info(f"📊 Sorted {len(all_artifacts_marked_for_deletion)} artifacts across all categories by size (largest first)")
        logger.info(f"💾 Total space to be freed: {format_size(total_size_to_delete)}")
        
        # Log the top 10 largest artifacts for visibility
        logger.info(f"🔝 Top 10 largest artifacts to be deleted:")
        for i, artifact in enumerate(all_artifacts_marked_for_deletion[:10]):
            tags_str = ", ".join(artifact['tags']) if artifact['tags'] else "none"
            logger.info(f"  {i+1:2d}. {artifact['digest'][:12]}... ({format_size(artifact['size'])}) - Category: {artifact['category']} - Tags: {tags_str}")
        
        if len(all_artifacts_marked_for_deletion) > 10:
            logger.info(f"  ... and {len(all_artifacts_marked_for_deletion) - 10} more artifacts")
    
    # Step 3: Execute deletions in parallel (if enabled and multiple artifacts to delete)
    if all_artifacts_marked_for_deletion:
        if ENABLE_PARALLEL and len(all_artifacts_marked_for_deletion) > 1:
            logger.info(f"🚀 Starting parallel deletion of {len(all_artifacts_marked_for_deletion)} artifacts using {min(ARTIFACT_WORKERS, len(all_artifacts_marked_for_deletion))} workers...")
            
            with ThreadPoolExecutor(max_workers=min(ARTIFACT_WORKERS, len(all_artifacts_marked_for_deletion))) as executor:
                # Submit all deletion tasks
                future_to_artifact = {
                    executor.submit(delete_artifact_parallel, project, repo, artifact_data): artifact_data
                    for artifact_data in all_artifacts_marked_for_deletion
                }
                
                # Process results as they complete
                for future in as_completed(future_to_artifact):
                    if shutdown_requested:
                        logger.warning("🚫 Shutdown requested, cancelling remaining deletions...")
                        break
                        
                    try:
                        result = future.result()
                        artifact_data = result['artifact_data']
                        digest = artifact_data['digest']
                        category = artifact_data['category']
                        
                        if result['success']:
                            total_space_to_cleanup += artifact_data['size']
                            category_stats[category]['to_delete'] += 1
                            category_stats[category]['delete_size'] += artifact_data['size']
                            artifacts_to_delete.append(artifact_data)
                            if result['actual_deletion']:
                                logger.debug(f"✅ Parallel deletion completed: {digest[:12]} ({category}) - {format_size(artifact_data['size'])}")
                        else:
                            # If deletion failed, treat as kept
                            error_msg = result.get('error', 'Unknown error')
                            logger.warning(f"⚠️  Failed to delete {digest[:12]} (parallel): {error_msg}")
                            artifacts_to_keep.append(artifact_data)
                    except Exception as e:
                        artifact_data = future_to_artifact[future]
                        digest = artifact_data['digest']
                        logger.error(f"❌ Exception in parallel deletion future for {digest[:12]}: {e}")
                        artifacts_to_keep.append(artifact_data)
            
            logger.info(f"🏁 Parallel deletion completed for all categories")
        else:
            # Sequential deletion for single artifact or when parallel is disabled
            logger.info(f"🔄 Processing {len(all_artifacts_marked_for_deletion)} artifact(s) sequentially...")
            
            for artifact_data in all_artifacts_marked_for_deletion:
                if shutdown_requested:
                    logger.warning("🚫 Shutdown requested, stopping artifact processing...")
                    break
                
                digest = artifact_data['digest']
                tags = artifact_data['tags']
                category = artifact_data['category']
                
                # Actually delete the artifact if not in dry run mode
                logger.debug(f"🎯 Sequential deletion: digest-based deletion only")
                deletion_success = delete_artifact_with_fallback(project, repo, digest, tags, artifact_data['artifact'])
                
                # Record progress for the global tracker
                global_progress_tracker.record_deletion(
                    success=deletion_success or DRY_RUN,
                    space_freed=artifact_data['size'] if (deletion_success or DRY_RUN) else 0
                )
                
                if deletion_success or DRY_RUN:
                    total_space_to_cleanup += artifact_data['size']
                    category_stats[category]['to_delete'] += 1
                    category_stats[category]['delete_size'] += artifact_data['size']
                    artifacts_to_delete.append(artifact_data)
                    if deletion_success:
                        logger.debug(f"✅ Sequential deletion completed: {digest[:12]} ({category}) - {format_size(artifact_data['size'])}")
                else:
                    # If deletion failed, treat as kept
                    logger.warning(f"⚠️  Failed to delete {digest[:12]}, treating as kept")
                    artifacts_to_keep.append(artifact_data)
    
    # Print summary for this repository
    logger.info(f"{'='*60}")
    run_mode = "(DRY RUN)" if DRY_RUN else "(LIVE RUN)"
    logger.info(f"🧹 CLEANUP SUMMARY FOR {project}/{repo} {run_mode}")
    logger.info(f"{'='*60}")
    logger.info(f"Total artifacts processed: {len(artifacts)}")
    logger.info(f"Artifacts to keep: {len(artifacts_to_keep)}")
    logger.info(f"Artifacts to delete: {len(artifacts_to_delete)}")
    
    if artifacts_to_delete:
        logger.info(f"📋 Artifacts marked for deletion:")
        for artifact in artifacts_to_delete:
            tags_str = ", ".join(artifact['tags']) if artifact['tags'] else "none"
            logger.info(f"  - {artifact['digest'][:12]}... ({format_size(artifact['size'])}) - Category: {artifact['category']} - Tags: {tags_str}")
    
    total_kept_space = sum(artifact['size'] for artifact in artifacts_to_keep)
    logger.info(f"💾 Space that will remain after cleanup: {format_size(total_kept_space)}")
    logger.info(f"💾 Space to be cleaned up: {format_size(total_space_to_cleanup)}")
    logger.info(f"💾 Total repository size: {format_size(total_space_to_cleanup + total_kept_space)}")
    
    if total_space_to_cleanup > 0:
        percentage_cleanup = (total_space_to_cleanup / (total_space_to_cleanup + total_kept_space)) * 100
        logger.info(f"🎯 Cleanup percentage: {percentage_cleanup:.1f}%")
    
    # Print category breakdown
    logger.info(f"📊 CLEANUP BREAKDOWN BY CATEGORY:")
    logger.info(f"============================================================")
    for category, stats in category_stats.items():
        if stats['total'] > 0:
            if stats['to_delete'] > 0:
                logger.info(f"🗑️  {stats['to_delete']} of {stats['total']} artifacts with {category} label to be cleaned up - which will free up {format_size(stats['delete_size'])}")
            else:
                logger.info(f"✅ 0 of {stats['total']} artifacts with {category} label to be cleaned up - all artifacts are within retention period")
    
    end_time = time.time()
    elapsed_time = end_time - start_time
    logger.info(f"⏱️  Repository {project}/{repo} processed in {format_duration(elapsed_time)}.")
    
    # Aggregate category totals for overall summary
    category_totals = {cat: stats['total'] for cat, stats in category_stats.items()}
    
    return {
        'repo': repo,
        'total_artifacts': len(artifacts),
        'artifacts_to_delete': len(artifacts_to_delete),
        'artifacts_to_keep': len(artifacts_to_keep),
        'space_to_cleanup': total_space_to_cleanup,
        'total_space': total_space_to_cleanup + total_kept_space,
        'processing_duration': elapsed_time,
        'category_totals': category_totals
    }

def process_repository_safe(project, repo_name, thread_id=None, artifact_count=None):
    """Thread-safe wrapper for process_repository with better output handling"""
    try:
        # Store thread ID for identification
        if thread_id is None:
            thread_id = threading.current_thread().ident
        
        # Capture output in a buffer to avoid mixed output between threads
        output_buffer = io.StringIO()
        error_buffer = io.StringIO()
        
        with redirect_stdout(output_buffer), redirect_stderr(error_buffer):
            result = process_repository(project, repo_name, artifact_count)
        
        # Get the captured output
        stdout_content = output_buffer.getvalue()
        stderr_content = error_buffer.getvalue()
        
        return {
            'success': True,
            'result': result,
            'stdout': stdout_content,
            'stderr': stderr_content,
            'thread_id': thread_id,
            'repo_name': repo_name
        }
        
    except Exception as e:
        return {
            'success': False,
            'error': str(e),
            'thread_id': thread_id,
            'repo_name': repo_name,
            'result': {
                'repo': repo_name[len(PROJECT)+1:] if repo_name.startswith(f"{PROJECT}/") else repo_name,
                'total_artifacts': 0,
                'artifacts_to_delete': 0,
                'artifacts_to_keep': 0,
                'space_to_cleanup': 0,
                'total_space': 0,
                'processing_duration': 0,
                'queue_time': 0,
                'total_time': 0
            }
        }

def get_selected_repositories():
    """Parse and validate selected repositories from REPO_LIST"""
    if not (REPO_LIST and USE_REPO_LIST):
        return None
        
    selected_repos = [repo.strip() for repo in REPO_LIST.split(',') if repo.strip()]
    
    if not selected_repos:
        logger.warning("⚠️  REPO_LIST is empty or contains only whitespace")
        logger.info("🔄 Falling back to processing ALL repositories")
        return None
    
    logger.info(f"🎯 Processing selected repositories: {', '.join(selected_repos)}")
    logger.info(f"   (USE_REPO_LIST=true, processing {len(selected_repos)} repositories)")
    
    return selected_repos

def prepare_repositories_with_info(selected_repos):
    """Get repository info for selected repositories and filter out empty ones"""
    repos_with_info = []
    skipped_repos = []
    
    # For selected repositories, use individual repository details API (more efficient)
    logger.info(f"🔍 Getting repository info for {len(selected_repos)} selected repositories...")
    
    for repo_name in selected_repos:
        # Use repository details API directly for each selected repo
        logger.debug(f"🔍 Fetching info for repository: {repo_name}")
        repo_info = get_repository_info(PROJECT, repo_name)
        
        if repo_info:
            artifact_count = repo_info.get('artifact_count', 0)
            if artifact_count == 0:
                logger.warning(f"⚠️  Skipping {repo_name} - no artifacts")
                skipped_repos.append((repo_name, "no artifacts"))
                continue
            elif artifact_count < MIN_ARTIFACTS_THRESHOLD:
                logger.info(f"⚠️  Skipping {repo_name} - only {artifact_count} artifacts (below threshold of {MIN_ARTIFACTS_THRESHOLD})")
                skipped_repos.append((repo_name, f"only {artifact_count} artifacts"))
                continue
            repos_with_info.append((repo_name, artifact_count))
        else:
            logger.warning(f"⚠️  Could not get artifact count for {repo_name}, processing without artifact count")
            repos_with_info.append((repo_name, None))
    
    if skipped_repos:
        logger.info(f"📊 Skipped {len(skipped_repos)} repositories:")
        for repo_name, reason in skipped_repos:
            logger.info(f"  - {repo_name}: {reason}")
    
    return repos_with_info

def process_repositories_parallel(repos_with_info):
    """Process multiple repositories in parallel"""
    results = []
    parallel_start = time.time()
    total_repos = len(repos_with_info)
    completed_count = 0
    
    logger.info(f"⚡ Using parallel processing with {min(MAX_WORKERS, total_repos)} workers")
    logger.info(f"⏱️  Parallel processing started at: {time.strftime('%H:%M:%S')}")
    
    # Log start times for each repository
    for repo_name, artifact_count in repos_with_info:
        logger.info(f"📦 Starting {repo_name} ({artifact_count} artifacts) at {time.strftime('%H:%M:%S')}")
    
    with ThreadPoolExecutor(max_workers=min(MAX_WORKERS, total_repos)) as executor:
        # Submit all repository processing tasks
        future_to_repo = {
            executor.submit(process_repository_safe, PROJECT, repo_name, artifact_count=artifact_count): (repo_name, artifact_count, time.time()) 
            for repo_name, artifact_count in repos_with_info
        }
        
        # Collect results as they complete
        for future in as_completed(future_to_repo):
            # Check for shutdown request
            if shutdown_requested:
                logger.warning("🚫 Shutdown requested, cancelling remaining tasks...")
                # Cancel remaining futures
                for f in future_to_repo:
                    f.cancel()
                break
                
            repo_name, artifact_count, submit_time = future_to_repo[future]
            completion_time = time.time()
            total_time = completion_time - submit_time
            completed_count += 1
            
            try:
                result = future.result()
                
                # Add timing metrics to the result
                actual_processing_time = result.get('result', {}).get('processing_duration', 0)
                queue_time = total_time - actual_processing_time if actual_processing_time > 0 else 0
                
                # Enhance result with timing metrics
                if 'result' in result:
                    result['result']['queue_time'] = queue_time
                    result['result']['total_time'] = total_time
                
                results.append(result)
                
                if result['success']:
                    # Get actual processing time from the repository result
                    actual_processing_time = result.get('result', {}).get('processing_duration', 0)
                    queue_time = total_time - actual_processing_time if actual_processing_time > 0 else 0
                    
                    # Format timing information
                    processing_display = format_duration(actual_processing_time)
                    queue_display = format_duration(queue_time)
                    total_display = format_duration(total_time)
                    
                    if queue_time > 1:  # Only show queue time if significant (>1 second)
                        time_info = f"(processing: {processing_display}, queue: {queue_display}, total: {total_display})"
                    else:
                        time_info = f"(took {processing_display})"
                    
                    # Calculate completion statistics
                    successful_repos = len([r for r in results if r.get('success', False)])
                    failed_repos = len([r for r in results if not r.get('success', True)])
                    remaining_repos = total_repos - completed_count
                    
                    # Calculate artifact statistics
                    total_artifacts_processed = sum(r['result'].get('total_artifacts', 0) for r in results if r.get('success', False))
                    total_artifacts_to_delete = sum(r['result'].get('artifacts_to_delete', 0) for r in results if r.get('success', False))
                    current_repo_artifacts = result.get('result', {}).get('total_artifacts', 0)
                    current_repo_to_delete = result.get('result', {}).get('artifacts_to_delete', 0)
                    
                    # Calculate space statistics
                    total_space_to_cleanup = sum(r['result'].get('space_to_cleanup', 0) for r in results if r.get('success', False))
                    total_space_processed = sum(r['result'].get('total_space', 0) for r in results if r.get('success', False))
                    current_repo_cleanup_space = result.get('result', {}).get('space_to_cleanup', 0)
                    current_repo_total_space = result.get('result', {}).get('total_space', 0)
                    
                    logger.info(f"✅ Completed {repo_name} at {time.strftime('%H:%M:%S')} {time_info}")
                    logger.info(f"   ├─ Processed {current_repo_artifacts:,} artifacts, marked {current_repo_to_delete:,} for deletion")
                    logger.info(f"   └─ Space: {format_size(current_repo_total_space)} total, {format_size(current_repo_cleanup_space)} to cleanup")
                    logger.info(f"📊 Progress: {completed_count}/{total_repos} repos done (✅ {successful_repos} success, ❌ {failed_repos} failed, 🔄 {remaining_repos} remaining)")
                    logger.info(f"📊 Artifacts: {total_artifacts_processed:,} processed, {total_artifacts_to_delete:,} marked for deletion across all completed repos")
                    logger.info(f"💾 Space: {format_size(total_space_processed)} processed, {format_size(total_space_to_cleanup)} to cleanup across all completed repos")
                else:
                    # Calculate completion statistics for failed case too
                    successful_repos = len([r for r in results if r.get('success', False)])
                    failed_repos = len([r for r in results if not r.get('success', True)])
                    remaining_repos = total_repos - completed_count
                    
                    # Calculate artifact statistics for failed case
                    total_artifacts_processed = sum(r['result'].get('total_artifacts', 0) for r in results if r.get('success', False))
                    total_artifacts_to_delete = sum(r['result'].get('artifacts_to_delete', 0) for r in results if r.get('success', False))
                    
                    # Calculate space statistics for failed case
                    total_space_to_cleanup = sum(r['result'].get('space_to_cleanup', 0) for r in results if r.get('success', False))
                    total_space_processed = sum(r['result'].get('total_space', 0) for r in results if r.get('success', False))
                    
                    logger.error(f"❌ Failed {repo_name} at {time.strftime('%H:%M:%S')} - {result['error']}")
                    logger.info(f"📊 Progress: {completed_count}/{total_repos} repos done (✅ {successful_repos} success, ❌ {failed_repos} failed, 🔄 {remaining_repos} remaining)")
                    logger.info(f"📊 Artifacts: {total_artifacts_processed:,} processed, {total_artifacts_to_delete:,} marked for deletion across successful repos")
                    logger.info(f"💾 Space: {format_size(total_space_processed)} processed, {format_size(total_space_to_cleanup)} to cleanup across successful repos")
                    
            except Exception as exc:
                # Calculate completion statistics for exception case too
                successful_repos = len([r for r in results if r.get('success', False)])
                failed_repos = len([r for r in results if not r.get('success', True)])
                remaining_repos = total_repos - completed_count
                
                # Calculate artifact statistics for exception case
                total_artifacts_processed = sum(r['result'].get('total_artifacts', 0) for r in results if r.get('success', False))
                total_artifacts_to_delete = sum(r['result'].get('artifacts_to_delete', 0) for r in results if r.get('success', False))
                
                # Calculate space statistics for exception case
                total_space_to_cleanup = sum(r['result'].get('space_to_cleanup', 0) for r in results if r.get('success', False))
                total_space_processed = sum(r['result'].get('total_space', 0) for r in results if r.get('success', False))
                
                logger.error(f"❌ Exception processing {repo_name}: {exc}")
                logger.info(f"📊 Progress: {completed_count}/{total_repos} repos done (✅ {successful_repos} success, ❌ {failed_repos} failed, 🔄 {remaining_repos} remaining)")
                logger.info(f"📊 Artifacts: {total_artifacts_processed:,} processed, {total_artifacts_to_delete:,} marked for deletion across successful repos")
                logger.info(f"💾 Space: {format_size(total_space_processed)} processed, {format_size(total_space_to_cleanup)} to cleanup across successful repos")
                results.append({
                    'success': False,
                    'error': str(exc),
                    'repo_name': repo_name,
                    'result': {
                        'repo': repo_name.split('/')[-1] if '/' in repo_name else repo_name,
                        'total_artifacts': 0,
                        'artifacts_to_delete': 0,
                        'artifacts_to_keep': 0,
                        'space_to_cleanup': 0,
                        'total_space': 0,
                        'processing_duration': 0,
                        'queue_time': 0,
                        'total_time': 0
                    }
                })
    
    parallel_end = time.time()
    logger.info(f"⏱️  Parallel processing completed in {format_duration(parallel_end - parallel_start)}")
    
    return results

def process_repositories_sequential(repos_with_info):
    """Process multiple repositories sequentially"""
    results = []
    total_repos = len(repos_with_info)
    completed_count = 0
    
    logger.info(f"🔄 Using sequential processing for {total_repos} repositories")
    for repo_name, artifact_count in repos_with_info:
        # Check for shutdown request
        if shutdown_requested:
            logger.warning(f"🚫 Shutdown requested, stopping before processing {repo_name}")
            break
            
        logger.info(f"📊 Processing selected repository: {repo_name} ({artifact_count if artifact_count else 'unknown'} artifacts)")
        result = process_repository(PROJECT, repo_name, artifact_count)
        completed_count += 1
        
        # For sequential processing, queue time is 0 and total time equals processing time
        processing_time = result.get('processing_duration', 0)
        result['queue_time'] = 0
        result['total_time'] = processing_time
        
        results.append({
            'success': True,
            'result': result,
            'repo_name': repo_name,
            'stdout': '',
            'stderr': ''
        })
        
        # Use actual processing time from the result
        processing_time = result.get('processing_duration', 0)
        time_info = f"(took {format_duration(processing_time)})" if processing_time > 0 else ""
        
        # Calculate completion statistics (for sequential, all are successful so far)
        successful_repos = completed_count  # All completed repos in sequential are successful
        failed_repos = 0  # No failures in sequential mode so far
        remaining_repos = total_repos - completed_count
        
        # Calculate artifact statistics for sequential processing
        total_artifacts_processed = sum(res['result'].get('total_artifacts', 0) for res in results)
        total_artifacts_to_delete = sum(res['result'].get('artifacts_to_delete', 0) for res in results)
        current_repo_artifacts = result.get('total_artifacts', 0)
        current_repo_to_delete = result.get('artifacts_to_delete', 0)
        
        # Calculate space statistics for sequential processing
        total_space_to_cleanup = sum(res['result'].get('space_to_cleanup', 0) for res in results)
        total_space_processed = sum(res['result'].get('total_space', 0) for res in results)
        current_repo_cleanup_space = result.get('space_to_cleanup', 0)
        current_repo_total_space = result.get('total_space', 0)
        
        logger.info(f"✅ Completed {repo_name} at {time.strftime('%H:%M:%S')} {time_info}")
        logger.info(f"   ├─ Processed {current_repo_artifacts:,} artifacts, marked {current_repo_to_delete:,} for deletion")
        logger.info(f"   └─ Space: {format_size(current_repo_total_space)} total, {format_size(current_repo_cleanup_space)} to cleanup")
        logger.info(f"📊 Progress: {completed_count}/{total_repos} repos done (✅ {successful_repos} success, ❌ {failed_repos} failed, 🔄 {remaining_repos} remaining)")
        logger.info(f"📊 Artifacts: {total_artifacts_processed:,} processed, {total_artifacts_to_delete:,} marked for deletion across all completed repos")
        logger.info(f"💾 Space: {format_size(total_space_processed)} processed, {format_size(total_space_to_cleanup)} to cleanup across all completed repos")
    
    return results

def process_selected_repositories(selected_repos):
    """Process a list of selected repositories"""
    repos_with_info = prepare_repositories_with_info(selected_repos)
    
    if ENABLE_PARALLEL and len(repos_with_info) > 1:
        return process_repositories_parallel(repos_with_info)
    else:
        return process_repositories_sequential(repos_with_info)

def process_all_repositories():
    """Process all repositories in the project"""
    logger.info(f"🎯 Processing ALL repositories in project: {PROJECT}")
    if REPO_LIST and not USE_REPO_LIST:
        logger.info("   (REPO_LIST ignored due to USE_REPO_LIST=false)")
    elif not REPO_LIST:
        logger.info("   (No REPO_LIST specified)")
    
    repo_fetch_start = time.time()
    from ..api.harbor_api import list_repositories
    repositories = list_repositories(PROJECT)
    repo_fetch_end = time.time()
    logger.info(f"⏱️  Repository list fetched in {format_duration(repo_fetch_end - repo_fetch_start)}")
    
    if not repositories:
        logger.warning("No repositories found to process.")
        return []
    
    # Filter repositories with artifacts
    repos_to_process = []
    skipped_repos = []
    
    for repo_data in repositories:
        repo_name = repo_data['name']
        artifact_count = repo_data.get('artifact_count', 0)
        
        if artifact_count == 0:
            logger.info(f"⚠️  Skipping {repo_name} - no artifacts")
            skipped_repos.append((repo_name, "no artifacts"))
            continue
        elif artifact_count < MIN_ARTIFACTS_THRESHOLD:
            logger.info(f"⚠️  Skipping {repo_name} - only {artifact_count} artifacts (below threshold of {MIN_ARTIFACTS_THRESHOLD})")
            skipped_repos.append((repo_name, f"only {artifact_count} artifacts"))
            continue

        repos_to_process.append((repo_name, artifact_count))
    
    logger.info(f"📊 Found {len(repos_to_process)} repositories with ≥{MIN_ARTIFACTS_THRESHOLD} artifacts to process")
    if skipped_repos:
        logger.info(f"📊 Skipped {len(skipped_repos)} repositories below threshold")
        from .config import LOG_LEVEL
        if LOG_LEVEL == "DEBUG":
            for repo_name, reason in skipped_repos:
                logger.info(f"  - {repo_name}: {reason}")
    
    if ENABLE_PARALLEL and len(repos_to_process) > 1:
        return process_repositories_parallel(repos_to_process)
    else:
        return process_repositories_sequential(repos_to_process) 