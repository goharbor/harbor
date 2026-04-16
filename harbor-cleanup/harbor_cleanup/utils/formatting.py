"""
Harbor Cleanup Tool - Utilities Module

This module contains common utility functions and helpers.
"""

import signal
import sys
import threading
from ..core.config import logger

# Global flag for graceful shutdown
shutdown_requested = False

def signal_handler(signum, frame):
    """Handle Ctrl+C gracefully"""
    global shutdown_requested
    shutdown_requested = True
    logger.warning("🚫 Ctrl+C detected! Requesting graceful shutdown...")
    logger.warning("⏱️  Waiting for current repository processing to complete...")
    logger.warning("🔥 Press Ctrl+C again for immediate force exit")
    
    # Set up second signal handler for force exit
    signal.signal(signal.SIGINT, force_exit_handler)

def force_exit_handler(signum, frame):
    """Force immediate exit on second Ctrl+C"""
    logger.error("🔥 Force exit requested! Terminating immediately...")
    sys.exit(1)

def setup_signal_handlers():
    """Set up signal handlers for graceful shutdown"""
    signal.signal(signal.SIGINT, signal_handler)

def format_size(size_bytes):
    """Convert bytes to human readable format (KB, MB, GB, TB)"""
    if size_bytes == 0:
        return "0 B"
    
    # Convert to appropriate unit
    if size_bytes < 1024:
        return f"{size_bytes} B"
    elif size_bytes < 1024 * 1024:
        return f"{size_bytes / 1024:.2f} KB"
    elif size_bytes < 1024 * 1024 * 1024:
        return f"{size_bytes / (1024 * 1024):.2f} MB"
    elif size_bytes < 1024 * 1024 * 1024 * 1024:
        return f"{size_bytes / (1024 * 1024 * 1024):.2f} GB"
    else:
        return f"{size_bytes / (1024 * 1024 * 1024 * 1024):.2f} TB"

def format_duration(seconds):
    """Convert seconds to human readable format (Xh Ym Zs)"""
    if seconds < 60:
        return f"{seconds:.1f}s"
    elif seconds < 3600:
        minutes = int(seconds // 60)
        remaining_seconds = seconds % 60
        return f"{minutes}m {remaining_seconds:.1f}s"
    else:
        hours = int(seconds // 3600)
        remaining_minutes = int((seconds % 3600) // 60)
        remaining_seconds = seconds % 60
        if remaining_seconds >= 1:
            return f"{hours}h {remaining_minutes}m {remaining_seconds:.1f}s"
        else:
            return f"{hours}h {remaining_minutes}m"

# Thread-local storage for printing (to avoid mixed output)
thread_local = threading.local() 