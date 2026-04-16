"""
Harbor Cleanup Tool - Checkpoint Manager

This module provides checkpoint functionality to resume interrupted cleanup operations.
"""

import json
import os
import time
from datetime import datetime
from typing import Dict, List, Optional, Set
from ..core.config import logger, PROJECT

class CheckpointManager:
    """Manages cleanup operation checkpoints for resumability"""
    
    def __init__(self):
        self.checkpoint_enabled = os.getenv("CHECKPOINT_ENABLED", "true").lower() in ("true", "1", "yes", "on")
        self.checkpoint_file = os.getenv("CHECKPOINT_FILE", f"harbor_cleanup_checkpoint_{PROJECT}.json")
        self.checkpoint_interval = int(os.getenv("CHECKPOINT_INTERVAL", "10"))  # Save every N repositories
        self.auto_resume = os.getenv("AUTO_RESUME", "false").lower() in ("true", "1", "yes", "on")
        
        self.current_session: Optional[Dict] = None
        self.processed_repositories: Set[str] = set()
        self.failed_repositories: Set[str] = set()
        self.repository_count = 0
        
    def initialize_session(self, all_repositories: List[str]) -> List[str]:
        """Initialize a new session or resume from checkpoint"""
        if not self.checkpoint_enabled:
            return all_repositories
            
        # Check for existing checkpoint
        checkpoint_data = self._load_checkpoint()
        
        if checkpoint_data and self.auto_resume:
            logger.info("🔄 Found existing checkpoint - attempting to resume...")
            return self._resume_from_checkpoint(checkpoint_data, all_repositories)
        elif checkpoint_data and not self.auto_resume:
            logger.info("🔄 Found existing checkpoint but AUTO_RESUME=false")
            logger.info("💡 Set AUTO_RESUME=true to resume automatically, or delete checkpoint file")
            logger.info(f"📁 Checkpoint file: {self.checkpoint_file}")
            return all_repositories
        else:
            logger.info("🆕 Starting new cleanup session...")
            return self._start_new_session(all_repositories)
    
    def _start_new_session(self, all_repositories: List[str]) -> List[str]:
        """Start a new cleanup session"""
        self.current_session = {
            'session_id': f"cleanup_{int(time.time())}",
            'start_time': time.time(),
            'project': PROJECT,
            'total_repositories': len(all_repositories),
            'all_repositories': all_repositories,
            'processed_repositories': [],
            'failed_repositories': [],
            'last_update': time.time()
        }
        
        self._save_checkpoint()
        logger.info(f"💾 Checkpoint system initialized for {len(all_repositories)} repositories")
        
        return all_repositories
    
    def _resume_from_checkpoint(self, checkpoint_data: Dict, all_repositories: List[str]) -> List[str]:
        """Resume cleanup from checkpoint"""
        self.current_session = checkpoint_data
        self.processed_repositories = set(checkpoint_data.get('processed_repositories', []))
        self.failed_repositories = set(checkpoint_data.get('failed_repositories', []))
        
        # Calculate remaining repositories
        remaining_repos = [repo for repo in all_repositories 
                          if repo not in self.processed_repositories]
        
        logger.info("📊 CHECKPOINT RESUME SUMMARY:")
        logger.info(f"   ├─ Session ID: {checkpoint_data.get('session_id', 'unknown')}")
        logger.info(f"   ├─ Original start: {self._format_timestamp(checkpoint_data.get('start_time', 0))}")
        logger.info(f"   ├─ Last update: {self._format_timestamp(checkpoint_data.get('last_update', 0))}")
        logger.info(f"   ├─ Total repositories: {len(all_repositories)}")
        logger.info(f"   ├─ Already processed: {len(self.processed_repositories)}")
        logger.info(f"   ├─ Failed repositories: {len(self.failed_repositories)}")
        logger.info(f"   └─ Remaining to process: {len(remaining_repos)}")
        
        if self.failed_repositories:
            logger.warning("⚠️  Previously failed repositories:")
            for repo in sorted(self.failed_repositories):
                logger.warning(f"     - {repo}")
        
        return remaining_repos
    
    def mark_repository_completed(self, repo_name: str, success: bool = True) -> None:
        """Mark a repository as completed (success or failure)"""
        if not self.checkpoint_enabled or not self.current_session:
            return
            
        if success:
            self.processed_repositories.add(repo_name)
            self.current_session['processed_repositories'] = list(self.processed_repositories)
        else:
            self.failed_repositories.add(repo_name)
            self.current_session['failed_repositories'] = list(self.failed_repositories)
        
        self.repository_count += 1
        self.current_session['last_update'] = time.time()
        
        # Save checkpoint periodically
        if self.repository_count % self.checkpoint_interval == 0:
            self._save_checkpoint()
            logger.debug(f"💾 Checkpoint saved after {self.repository_count} repositories")
    
    def finalize_session(self) -> None:
        """Finalize the cleanup session and cleanup checkpoint"""
        if not self.checkpoint_enabled:
            return
            
        if self.current_session:
            self.current_session['completed'] = True
            self.current_session['end_time'] = time.time()
            self._save_checkpoint()
            
            logger.info("✅ Cleanup session completed successfully")
            
            # Clean up checkpoint file after successful completion
            self._cleanup_checkpoint()
    
    def mark_interrupted(self) -> None:
        """Mark session as interrupted and save final checkpoint"""
        if not self.checkpoint_enabled or not self.current_session:
            return
            
        self.current_session['interrupted'] = True
        self.current_session['interrupt_time'] = time.time()
        self._save_checkpoint()
        
        logger.warning("🚫 Session interrupted - checkpoint saved for resume")
        logger.info(f"💡 To resume: Set AUTO_RESUME=true and restart the cleanup")
        logger.info(f"📁 Checkpoint file: {self.checkpoint_file}")
    
    def _save_checkpoint(self) -> None:
        """Save current checkpoint to file"""
        try:
            checkpoint_data = {
                **self.current_session,
                'checkpoint_version': '1.0',
                'saved_at': datetime.now().isoformat()
            }
            
            # Write to temporary file first, then rename (atomic operation)
            temp_file = f"{self.checkpoint_file}.tmp"
            with open(temp_file, 'w') as f:
                json.dump(checkpoint_data, f, indent=2)
            
            os.rename(temp_file, self.checkpoint_file)
            logger.debug(f"💾 Checkpoint saved: {self.checkpoint_file}")
            
        except Exception as e:
            logger.error(f"❌ Failed to save checkpoint: {e}")
    
    def _load_checkpoint(self) -> Optional[Dict]:
        """Load checkpoint from file"""
        try:
            if not os.path.exists(self.checkpoint_file):
                return None
                
            with open(self.checkpoint_file, 'r') as f:
                data = json.load(f)
                
            # Validate checkpoint data
            if not self._validate_checkpoint(data):
                logger.warning("⚠️  Invalid checkpoint data - starting fresh")
                return None
                
            return data
            
        except Exception as e:
            logger.error(f"❌ Failed to load checkpoint: {e}")
            return None
    
    def _validate_checkpoint(self, data: Dict) -> bool:
        """Validate checkpoint data structure"""
        required_fields = ['session_id', 'start_time', 'project', 'all_repositories']
        return all(field in data for field in required_fields)
    
    def _cleanup_checkpoint(self) -> None:
        """Remove checkpoint file after successful completion"""
        try:
            if os.path.exists(self.checkpoint_file):
                os.remove(self.checkpoint_file)
                logger.info(f"🗑️  Checkpoint file cleaned up: {self.checkpoint_file}")
        except Exception as e:
            logger.warning(f"⚠️  Failed to cleanup checkpoint file: {e}")
    
    def _format_timestamp(self, timestamp: float) -> str:
        """Format timestamp for display"""
        if timestamp == 0:
            return "unknown"
        return datetime.fromtimestamp(timestamp).strftime("%Y-%m-%d %H:%M:%S")
    
    def get_resume_info(self) -> Optional[Dict]:
        """Get information about resumable session"""
        checkpoint_data = self._load_checkpoint()
        if not checkpoint_data:
            return None
            
        processed = len(checkpoint_data.get('processed_repositories', []))
        failed = len(checkpoint_data.get('failed_repositories', []))
        total = checkpoint_data.get('total_repositories', 0)
        remaining = total - processed - failed
        
        return {
            'session_id': checkpoint_data.get('session_id'),
            'start_time': checkpoint_data.get('start_time'),
            'last_update': checkpoint_data.get('last_update'),
            'total_repositories': total,
            'processed_repositories': processed,
            'failed_repositories': failed,
            'remaining_repositories': remaining,
            'completion_percentage': (processed / total * 100) if total > 0 else 0,
            'checkpoint_file': self.checkpoint_file
        }

# Global checkpoint manager instance
checkpoint_manager = CheckpointManager() 