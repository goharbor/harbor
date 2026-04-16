"""
Harbor Cleanup Tool - Logging Utilities

This module provides enhanced logging utilities for better structured output.
"""

import time
import functools
from ..core.config import logger, PERFORMANCE_LOGGING, VERBOSE_ARTIFACTS

class DeletionProgressTracker:
    """Tracks deletion progress and provides periodic summaries"""
    
    def __init__(self, summary_interval=100):
        self.summary_interval = summary_interval
        self.total_deletions = 0
        self.successful_deletions = 0
        self.failed_deletions = 0
        self.total_space_freed = 0
        self.start_time = time.time()
        self.last_summary_count = 0
        
    def record_deletion(self, success, space_freed=0):
        """Record a deletion attempt"""
        self.total_deletions += 1
        if success:
            self.successful_deletions += 1
            self.total_space_freed += space_freed
        else:
            self.failed_deletions += 1
            
        # Check if we should print a summary
        if self.total_deletions % self.summary_interval == 0:
            self._print_progress_summary()
            
    def _print_progress_summary(self):
        """Print a progress summary"""
        from .formatting import format_size, format_duration
        
        elapsed_time = time.time() - self.start_time
        deletions_this_batch = self.total_deletions - self.last_summary_count
        
        # Calculate rates
        avg_deletions_per_sec = self.total_deletions / elapsed_time if elapsed_time > 0 else 0
        
        logger.info(f"📊 DELETION PROGRESS SUMMARY (after {self.total_deletions} attempts)")
        logger.info(f"   ├─ Successful: {self.successful_deletions:,}")
        logger.info(f"   ├─ Failed: {self.failed_deletions:,}")
        logger.info(f"   ├─ Success rate: {(self.successful_deletions/self.total_deletions*100):.1f}%" if self.total_deletions > 0 else "   ├─ Success rate: 0%")
        logger.info(f"   ├─ Space freed: {format_size(self.total_space_freed)}")
        logger.info(f"   ├─ Elapsed time: {format_duration(elapsed_time)}")
        logger.info(f"   └─ Average rate: {avg_deletions_per_sec:.1f} deletions/sec")
        
        self.last_summary_count = self.total_deletions
        
    def get_final_summary(self):
        """Get final summary stats"""
        from .formatting import format_size, format_duration
        
        elapsed_time = time.time() - self.start_time
        return {
            'total_deletions': self.total_deletions,
            'successful_deletions': self.successful_deletions,
            'failed_deletions': self.failed_deletions,
            'total_space_freed': self.total_space_freed,
            'total_space_freed_formatted': format_size(self.total_space_freed),
            'elapsed_time': elapsed_time,
            'elapsed_time_formatted': format_duration(elapsed_time),
            'success_rate': (self.successful_deletions/self.total_deletions*100) if self.total_deletions > 0 else 0,
            'avg_rate': self.total_deletions / elapsed_time if elapsed_time > 0 else 0
        }

def log_performance(func):
    """Decorator to log function performance metrics"""
    @functools.wraps(func)
    def wrapper(*args, **kwargs):
        if not PERFORMANCE_LOGGING:
            return func(*args, **kwargs)
        
        start_time = time.time()
        func_name = func.__name__
        
        try:
            logger.debug(f"⚡ START: {func_name}")
            result = func(*args, **kwargs)
            elapsed = time.time() - start_time
            logger.debug(f"⚡ DONE: {func_name} ({elapsed:.2f}s)")
            return result
        except Exception as e:
            elapsed = time.time() - start_time
            logger.error(f"⚡ ERROR: {func_name} failed after {elapsed:.2f}s: {e}")
            raise
    
    return wrapper

def log_api_call(method, url, status_code, elapsed_time=None):
    """Log API call with consistent formatting"""
    emoji = "✅" if 200 <= status_code < 300 else "⚠️" if 400 <= status_code < 500 else "❌"
    timing = f" ({elapsed_time:.2f}s)" if elapsed_time else ""
    logger.debug(f"🌐 {emoji} {method} {status_code}{timing}")

def log_artifact_action(action, artifact_ref, repo=None, details=None):
    """Log artifact actions with consistent formatting"""
    repo_info = f" in {repo}" if repo else ""
    detail_info = f" - {details}" if details else ""
    
    emoji_map = {
        'keep': '✅',
        'delete': '🗑️',
        'skip': '⏭️',
        'error': '❌',
        'processing': '🔄'
    }
    
    emoji = emoji_map.get(action.lower(), '📋')
    logger.info(f"{emoji} {action.upper()}: {artifact_ref}{repo_info}{detail_info}")

def log_section_start(title, level="info"):
    """Log section start with consistent formatting"""
    separator = "=" * 60
    log_func = getattr(logger, level.lower(), logger.info)
    log_func(separator)
    log_func(f"📋 {title.upper()}")
    log_func(separator)

def log_section_end(title=None):
    """Log section end"""
    if title:
        logger.info(f"✅ {title} completed")
    logger.info("")

def log_parallel_progress(current, total, operation="Processing"):
    """Log parallel processing progress"""
    percentage = (current / total * 100) if total > 0 else 0
    logger.info(f"📊 {operation} progress: {current}/{total} ({percentage:.1f}%)")

def log_repository_summary(repo_name, artifacts_processed, artifacts_deleted, space_freed, duration):
    """Log repository processing summary with consistent format"""
    from .formatting import format_size, format_duration
    
    logger.info(f"📊 Repository Summary: {repo_name}")
    logger.info(f"   ├─ Artifacts processed: {artifacts_processed:,}")
    logger.info(f"   ├─ Artifacts deleted: {artifacts_deleted:,}")
    logger.info(f"   ├─ Space freed: {format_size(space_freed)}")
    logger.info(f"   └─ Duration: {format_duration(duration)}")

def log_error_context(error_type, context, suggestion=None):
    """Log errors with helpful context and suggestions"""
    logger.error(f"❌ {error_type}")
    logger.error(f"   📋 Context: {context}")
    if suggestion:
        logger.error(f"   💡 Suggestion: {suggestion}")

def conditional_debug(condition, message):
    """Log debug message only if condition is true"""
    if condition and VERBOSE_ARTIFACTS:
        logger.debug(message)

class LogContext:
    """Context manager for grouped logging"""
    
    def __init__(self, operation, level="info"):
        self.operation = operation
        self.level = level
        self.start_time = None
    
    def __enter__(self):
        self.start_time = time.time()
        log_func = getattr(logger, self.level.lower(), logger.info)
        log_func(f"🚀 Starting: {self.operation}")
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        duration = time.time() - self.start_time if self.start_time else 0
        
        if exc_type is None:
            logger.info(f"✅ Completed: {self.operation} ({duration:.2f}s)")
        else:
            logger.error(f"❌ Failed: {self.operation} after {duration:.2f}s")
            logger.error(f"   Error: {exc_val}")
        
        return False  # Don't suppress exceptions 