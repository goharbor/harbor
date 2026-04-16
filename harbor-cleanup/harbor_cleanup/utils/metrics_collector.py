"""
Harbor Cleanup Tool - Metrics Collection

This module provides comprehensive metrics collection for monitoring,
analytics, and performance optimization.
"""

import time
import json
import os
from datetime import datetime, timezone
from dataclasses import dataclass, asdict
from typing import Dict, List, Optional
from ..core.config import logger, PROJECT, DRY_RUN

@dataclass
class RepositoryMetrics:
    """Metrics for a single repository cleanup"""
    name: str
    start_time: float
    end_time: float
    artifacts_total: int
    artifacts_deleted: int
    artifacts_kept: int
    artifacts_failed: int
    space_total: int
    space_deleted: int
    space_kept: int
    categories: Dict[str, int]  # Category breakdown
    deletion_failures: List[Dict]  # Failed deletion details
    
    @property
    def duration(self) -> float:
        return self.end_time - self.start_time
    
    @property
    def success_rate(self) -> float:
        if self.artifacts_total == 0:
            return 0.0
        return (self.artifacts_deleted / self.artifacts_total) * 100

@dataclass
class CleanupSessionMetrics:
    """Metrics for entire cleanup session"""
    session_id: str
    start_time: float
    end_time: Optional[float] = None
    project: str = PROJECT
    dry_run: bool = DRY_RUN
    repositories: List[RepositoryMetrics] = None
    harbor_api_calls: int = 0
    harbor_api_errors: int = 0
    harbor_api_retries: int = 0
    gc_triggered: bool = False
    gc_success: bool = False
    interrupted: bool = False
    
    def __post_init__(self):
        if self.repositories is None:
            self.repositories = []
    
    @property
    def duration(self) -> float:
        if self.end_time is None:
            return time.time() - self.start_time
        return self.end_time - self.start_time
    
    @property
    def total_artifacts_processed(self) -> int:
        return sum(repo.artifacts_total for repo in self.repositories)
    
    @property
    def total_artifacts_deleted(self) -> int:
        return sum(repo.artifacts_deleted for repo in self.repositories)
    
    @property
    def total_space_freed(self) -> int:
        return sum(repo.space_deleted for repo in self.repositories)
    
    @property
    def overall_success_rate(self) -> float:
        total = self.total_artifacts_processed
        if total == 0:
            return 0.0
        return (self.total_artifacts_deleted / total) * 100

class MetricsCollector:
    """Collects and manages cleanup metrics"""
    
    def __init__(self):
        self.session = CleanupSessionMetrics(
            session_id=f"cleanup_{int(time.time())}",
            start_time=time.time()
        )
        self.current_repo_metrics: Optional[RepositoryMetrics] = None
        self.metrics_enabled = os.getenv("METRICS_ENABLED", "true").lower() in ("true", "1", "yes", "on")
        self.metrics_output_file = os.getenv("METRICS_OUTPUT_FILE", "harbor_cleanup_metrics.json")
        
    def start_repository(self, repo_name: str) -> None:
        """Start metrics collection for a repository"""
        if not self.metrics_enabled:
            return
            
        self.current_repo_metrics = RepositoryMetrics(
            name=repo_name,
            start_time=time.time(),
            end_time=0.0,
            artifacts_total=0,
            artifacts_deleted=0,
            artifacts_kept=0,
            artifacts_failed=0,
            space_total=0,
            space_deleted=0,
            space_kept=0,
            categories={},
            deletion_failures=[]
        )
        logger.debug(f"📊 Started metrics collection for repository: {repo_name}")
    
    def record_artifact_processed(self, artifact_data: Dict, action: str, success: bool = True, error: str = None) -> None:
        """Record artifact processing metrics"""
        if not self.metrics_enabled or not self.current_repo_metrics:
            return
            
        size = artifact_data.get('size', 0)
        category = artifact_data.get('category', 'unknown')
        
        self.current_repo_metrics.artifacts_total += 1
        self.current_repo_metrics.space_total += size
        
        # Update category counts
        if category not in self.current_repo_metrics.categories:
            self.current_repo_metrics.categories[category] = 0
        self.current_repo_metrics.categories[category] += 1
        
        if action == 'delete':
            if success:
                self.current_repo_metrics.artifacts_deleted += 1
                self.current_repo_metrics.space_deleted += size
            else:
                self.current_repo_metrics.artifacts_failed += 1
                if error:
                    self.current_repo_metrics.deletion_failures.append({
                        'digest': artifact_data.get('digest', 'unknown')[:12],
                        'tags': artifact_data.get('tags', []),
                        'error': error,
                        'timestamp': time.time()
                    })
        elif action == 'keep':
            self.current_repo_metrics.artifacts_kept += 1
            self.current_repo_metrics.space_kept += size
    
    def end_repository(self) -> None:
        """Finish metrics collection for current repository"""
        if not self.metrics_enabled or not self.current_repo_metrics:
            return
            
        self.current_repo_metrics.end_time = time.time()
        self.session.repositories.append(self.current_repo_metrics)
        
        # Log repository summary
        repo = self.current_repo_metrics
        logger.info(f"📊 Repository Metrics: {repo.name}")
        logger.info(f"   ├─ Duration: {repo.duration:.2f}s")
        logger.info(f"   ├─ Artifacts: {repo.artifacts_total:,} total, {repo.artifacts_deleted:,} deleted")
        logger.info(f"   ├─ Success rate: {repo.success_rate:.1f}%")
        logger.info(f"   └─ Space freed: {self._format_size(repo.space_deleted)}")
        
        self.current_repo_metrics = None
    
    def record_api_call(self, success: bool = True, retry: bool = False) -> None:
        """Record Harbor API call metrics"""
        if not self.metrics_enabled:
            return
            
        self.session.harbor_api_calls += 1
        if not success:
            self.session.harbor_api_errors += 1
        if retry:
            self.session.harbor_api_retries += 1
    
    def record_gc_trigger(self, success: bool) -> None:
        """Record garbage collection trigger"""
        if not self.metrics_enabled:
            return
            
        self.session.gc_triggered = True
        self.session.gc_success = success
    
    def mark_interrupted(self) -> None:
        """Mark session as interrupted"""
        if not self.metrics_enabled:
            return
            
        self.session.interrupted = True
    
    def finalize_session(self) -> None:
        """Finalize the cleanup session metrics"""
        if not self.metrics_enabled:
            return
            
        self.session.end_time = time.time()
        
        # Log session summary
        logger.info(f"📊 SESSION METRICS SUMMARY:")
        logger.info(f"   ├─ Session ID: {self.session.session_id}")
        logger.info(f"   ├─ Duration: {self.session.duration:.2f}s")
        logger.info(f"   ├─ Repositories processed: {len(self.session.repositories)}")
        logger.info(f"   ├─ Total artifacts: {self.session.total_artifacts_processed:,}")
        logger.info(f"   ├─ Total deleted: {self.session.total_artifacts_deleted:,}")
        logger.info(f"   ├─ Overall success rate: {self.session.overall_success_rate:.1f}%")
        logger.info(f"   ├─ Total space freed: {self._format_size(self.session.total_space_freed)}")
        logger.info(f"   ├─ API calls: {self.session.harbor_api_calls:,} (errors: {self.session.harbor_api_errors:,}, retries: {self.session.harbor_api_retries:,})")
        logger.info(f"   └─ GC triggered: {'✅' if self.session.gc_triggered else '❌'}")
        
        # Save metrics to file
        self._save_metrics()
    
    def _save_metrics(self) -> None:
        """Save metrics to JSON file"""
        try:
            metrics_data = {
                'session': asdict(self.session),
                'generated_at': datetime.now(timezone.utc).isoformat(),
                'version': '1.0'
            }
            
            with open(self.metrics_output_file, 'w') as f:
                json.dump(metrics_data, f, indent=2, default=str)
            
            logger.info(f"📊 Metrics saved to: {self.metrics_output_file}")
            
        except Exception as e:
            logger.error(f"❌ Failed to save metrics: {e}")
    
    def _format_size(self, size_bytes: int) -> str:
        """Format size in human readable format"""
        from ..utils.formatting import format_size
        return format_size(size_bytes)
    
    def get_current_stats(self) -> Dict:
        """Get current session statistics"""
        return {
            'session_id': self.session.session_id,
            'duration': self.session.duration,
            'repositories_processed': len(self.session.repositories),
            'total_artifacts': self.session.total_artifacts_processed,
            'total_deleted': self.session.total_artifacts_deleted,
            'success_rate': self.session.overall_success_rate,
            'space_freed': self.session.total_space_freed,
            'api_calls': self.session.harbor_api_calls,
            'api_errors': self.session.harbor_api_errors
        }

# Global metrics collector instance
metrics_collector = MetricsCollector() 