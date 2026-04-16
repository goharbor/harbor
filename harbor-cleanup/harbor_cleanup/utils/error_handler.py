"""
Harbor Cleanup Tool - Error Handling & Recovery

This module provides comprehensive error handling, retry mechanisms,
and recovery strategies for robust cleanup operations.
"""

import time
import functools
from typing import Dict, List, Optional, Callable, Any
from enum import Enum
from dataclasses import dataclass
from ..core.config import logger

class ErrorSeverity(Enum):
    """Error severity levels"""
    LOW = "low"           # Minor issues that don't affect operation
    MEDIUM = "medium"     # Issues that might affect performance
    HIGH = "high"         # Issues that affect functionality
    CRITICAL = "critical" # Issues that stop operation

class ErrorCategory(Enum):
    """Error categories for better handling"""
    NETWORK = "network"           # Network connectivity issues
    AUTHENTICATION = "auth"       # Authentication/authorization issues  
    API_RATE_LIMIT = "rate_limit" # API rate limiting
    HARBOR_SERVER = "harbor"      # Harbor server errors
    PERMISSION = "permission"     # Permission denied errors
    NOT_FOUND = "not_found"      # Resource not found
    DEPENDENCY = "dependency"     # Missing dependencies
    CONFIGURATION = "config"      # Configuration errors
    UNKNOWN = "unknown"          # Unknown errors

@dataclass
class ErrorInfo:
    """Structured error information"""
    category: ErrorCategory
    severity: ErrorSeverity
    message: str
    details: Optional[str] = None
    suggestion: Optional[str] = None
    retry_after: Optional[float] = None
    
class ErrorClassifier:
    """Classifies errors for appropriate handling"""
    
    ERROR_PATTERNS = {
        # Network errors
        ErrorCategory.NETWORK: [
            "connection timeout", "connection refused", "network unreachable",
            "dns resolution failed", "connection reset", "ssl error"
        ],
        
        # Authentication errors
        ErrorCategory.AUTHENTICATION: [
            "unauthorized", "authentication failed", "invalid credentials",
            "access denied", "permission denied", "forbidden"
        ],
        
        # Rate limiting
        ErrorCategory.API_RATE_LIMIT: [
            "rate limit", "too many requests", "quota exceeded",
            "throttling", "api limit"
        ],
        
        # Harbor specific errors
        ErrorCategory.HARBOR_SERVER: [
            "internal server error", "harbor error", "database error",
            "storage error", "registry error"
        ],
        
        # Permission errors
        ErrorCategory.PERMISSION: [
            "permission denied", "access forbidden", "insufficient privileges"
        ],
        
        # Not found errors
        ErrorCategory.NOT_FOUND: [
            "not found", "does not exist", "unknown repository",
            "artifact not found", "404"
        ]
    }
    
    @classmethod
    def classify_error(cls, error_message: str, status_code: Optional[int] = None) -> ErrorInfo:
        """Classify an error and provide structured information"""
        error_message_lower = error_message.lower()
        
        # Check status code first
        if status_code:
            if status_code == 401:
                return ErrorInfo(
                    category=ErrorCategory.AUTHENTICATION,
                    severity=ErrorSeverity.HIGH,
                    message=error_message,
                    suggestion="Check Harbor credentials (HARBOR_USERNAME/HARBOR_PASSWORD)"
                )
            elif status_code == 403:
                return ErrorInfo(
                    category=ErrorCategory.PERMISSION,
                    severity=ErrorSeverity.HIGH,
                    message=error_message,
                    suggestion="Check Harbor user permissions for repository access"
                )
            elif status_code == 404:
                return ErrorInfo(
                    category=ErrorCategory.NOT_FOUND,
                    severity=ErrorSeverity.MEDIUM,
                    message=error_message,
                    suggestion="Repository or artifact may have been deleted"
                )
            elif status_code == 429:
                return ErrorInfo(
                    category=ErrorCategory.API_RATE_LIMIT,
                    severity=ErrorSeverity.MEDIUM,
                    message=error_message,
                    suggestion="Reduce parallelism or add delays",
                    retry_after=60.0
                )
            elif status_code >= 500:
                return ErrorInfo(
                    category=ErrorCategory.HARBOR_SERVER,
                    severity=ErrorSeverity.HIGH,
                    message=error_message,
                    suggestion="Harbor server issue - check Harbor logs"
                )
        
        # Check error message patterns
        for category, patterns in cls.ERROR_PATTERNS.items():
            if any(pattern in error_message_lower for pattern in patterns):
                severity = cls._get_severity_for_category(category)
                suggestion = cls._get_suggestion_for_category(category)
                retry_after = cls._get_retry_delay_for_category(category)
                
                return ErrorInfo(
                    category=category,
                    severity=severity,
                    message=error_message,
                    suggestion=suggestion,
                    retry_after=retry_after
                )
        
        # Default classification
        return ErrorInfo(
            category=ErrorCategory.UNKNOWN,
            severity=ErrorSeverity.MEDIUM,
            message=error_message,
            suggestion="Check logs for more details"
        )
    
    @staticmethod
    def _get_severity_for_category(category: ErrorCategory) -> ErrorSeverity:
        """Get default severity for error category"""
        severity_map = {
            ErrorCategory.NETWORK: ErrorSeverity.HIGH,
            ErrorCategory.AUTHENTICATION: ErrorSeverity.CRITICAL,
            ErrorCategory.API_RATE_LIMIT: ErrorSeverity.MEDIUM,
            ErrorCategory.HARBOR_SERVER: ErrorSeverity.HIGH,
            ErrorCategory.PERMISSION: ErrorSeverity.HIGH,
            ErrorCategory.NOT_FOUND: ErrorSeverity.LOW,
            ErrorCategory.DEPENDENCY: ErrorSeverity.CRITICAL,
            ErrorCategory.CONFIGURATION: ErrorSeverity.CRITICAL
        }
        return severity_map.get(category, ErrorSeverity.MEDIUM)
    
    @staticmethod
    def _get_suggestion_for_category(category: ErrorCategory) -> str:
        """Get suggestion for error category"""
        suggestions = {
            ErrorCategory.NETWORK: "Check network connectivity and Harbor URL",
            ErrorCategory.AUTHENTICATION: "Verify Harbor credentials",
            ErrorCategory.API_RATE_LIMIT: "Reduce parallelism or add delays",
            ErrorCategory.HARBOR_SERVER: "Check Harbor server status and logs",
            ErrorCategory.PERMISSION: "Check Harbor user permissions",
            ErrorCategory.NOT_FOUND: "Resource may have been deleted or moved",
            ErrorCategory.DEPENDENCY: "Install missing dependencies",
            ErrorCategory.CONFIGURATION: "Check configuration settings"
        }
        return suggestions.get(category, "Review logs for more information")
    
    @staticmethod
    def _get_retry_delay_for_category(category: ErrorCategory) -> Optional[float]:
        """Get retry delay for error category"""
        delays = {
            ErrorCategory.NETWORK: 5.0,
            ErrorCategory.API_RATE_LIMIT: 60.0,
            ErrorCategory.HARBOR_SERVER: 10.0,
            ErrorCategory.NOT_FOUND: None  # Don't retry not found errors
        }
        return delays.get(category, 2.0)

class ErrorRecoveryStrategy:
    """Provides recovery strategies for different error types"""
    
    def __init__(self):
        self.error_count_by_category: Dict[ErrorCategory, int] = {}
        self.consecutive_errors = 0
        self.max_consecutive_errors = 10
        
    def should_retry(self, error_info: ErrorInfo) -> bool:
        """Determine if operation should be retried"""
        # Never retry critical errors
        if error_info.severity == ErrorSeverity.CRITICAL:
            return False
        
        # Don't retry not found errors
        if error_info.category == ErrorCategory.NOT_FOUND:
            return False
        
        # Check consecutive error limit
        if self.consecutive_errors >= self.max_consecutive_errors:
            logger.warning(f"⚠️  Too many consecutive errors ({self.consecutive_errors}), stopping retries")
            return False
        
        # Category-specific retry logic
        category_count = self.error_count_by_category.get(error_info.category, 0)
        
        if error_info.category == ErrorCategory.API_RATE_LIMIT:
            return category_count < 3  # Max 3 rate limit retries
        elif error_info.category == ErrorCategory.HARBOR_SERVER:
            return category_count < 5  # Max 5 server error retries
        elif error_info.category == ErrorCategory.NETWORK:
            return category_count < 3  # Max 3 network retries
        
        return category_count < 2  # Default max 2 retries
    
    def get_retry_delay(self, error_info: ErrorInfo, attempt: int) -> float:
        """Get delay before retry attempt"""
        base_delay = error_info.retry_after or 1.0
        
        # Exponential backoff
        delay = base_delay * (2 ** (attempt - 1))
        
        # Cap maximum delay
        max_delay = 60.0 if error_info.category == ErrorCategory.API_RATE_LIMIT else 30.0
        
        return min(delay, max_delay)
    
    def record_error(self, error_info: ErrorInfo) -> None:
        """Record error for tracking"""
        self.error_count_by_category[error_info.category] = \
            self.error_count_by_category.get(error_info.category, 0) + 1
        self.consecutive_errors += 1
    
    def record_success(self) -> None:
        """Record successful operation"""
        self.consecutive_errors = 0
    
    def suggest_recovery_action(self, error_info: ErrorInfo) -> Optional[str]:
        """Suggest recovery action for error"""
        if error_info.category == ErrorCategory.API_RATE_LIMIT:
            return "Consider reducing MAX_WORKERS and enabling longer API_RETRY_DELAY"
        elif error_info.category == ErrorCategory.HARBOR_SERVER:
            return "Check Harbor server health and consider running during off-peak hours"
        elif error_info.category == ErrorCategory.AUTHENTICATION:
            return "Verify Harbor credentials and check user account status"
        elif error_info.category == ErrorCategory.NETWORK:
            return "Check network connectivity and Harbor URL accessibility"
        
        return error_info.suggestion

def with_error_handling(
    operation_name: str,
    max_retries: int = 3,
    raise_on_error: bool = True
):
    """Decorator for comprehensive error handling"""
    def decorator(func: Callable) -> Callable:
        @functools.wraps(func)
        def wrapper(*args, **kwargs) -> Any:
            recovery_strategy = ErrorRecoveryStrategy()
            
            for attempt in range(1, max_retries + 2):  # +1 for initial attempt
                try:
                    result = func(*args, **kwargs)
                    recovery_strategy.record_success()
                    return result
                    
                except Exception as e:
                    # Classify the error
                    status_code = getattr(e, 'status_code', None)
                    error_info = ErrorClassifier.classify_error(str(e), status_code)
                    
                    # Record error
                    recovery_strategy.record_error(error_info)
                    
                    # Log structured error information
                    logger.error(f"❌ {operation_name} failed (attempt {attempt}/{max_retries + 1})")
                    logger.error(f"   📋 Category: {error_info.category.value}")
                    logger.error(f"   ⚠️  Severity: {error_info.severity.value}")
                    logger.error(f"   💬 Message: {error_info.message}")
                    
                    if error_info.suggestion:
                        logger.error(f"   💡 Suggestion: {error_info.suggestion}")
                    
                    # Check if we should retry
                    if attempt <= max_retries and recovery_strategy.should_retry(error_info):
                        delay = recovery_strategy.get_retry_delay(error_info, attempt)
                        logger.warning(f"🔄 Retrying {operation_name} in {delay:.1f}s...")
                        time.sleep(delay)
                        continue
                    else:
                        # Final attempt failed
                        recovery_action = recovery_strategy.suggest_recovery_action(error_info)
                        if recovery_action:
                            logger.error(f"   🔧 Recovery: {recovery_action}")
                        
                        if raise_on_error:
                            raise
                        else:
                            return None
            
            return None  # Should never reach here
        
        return wrapper
    return decorator

class CircuitBreaker:
    """Circuit breaker pattern for failing operations"""
    
    def __init__(self, failure_threshold: int = 5, recovery_timeout: float = 60.0):
        self.failure_threshold = failure_threshold
        self.recovery_timeout = recovery_timeout
        self.failure_count = 0
        self.last_failure_time = 0
        self.state = "CLOSED"  # CLOSED, OPEN, HALF_OPEN
    
    def call(self, func: Callable, *args, **kwargs) -> Any:
        """Execute function with circuit breaker protection"""
        if self.state == "OPEN":
            if time.time() - self.last_failure_time > self.recovery_timeout:
                self.state = "HALF_OPEN"
                logger.info("🔄 Circuit breaker attempting recovery...")
            else:
                raise Exception("Circuit breaker is OPEN - operation blocked")
        
        try:
            result = func(*args, **kwargs)
            
            if self.state == "HALF_OPEN":
                self.state = "CLOSED"
                self.failure_count = 0
                logger.info("✅ Circuit breaker recovered - state is CLOSED")
            
            return result
            
        except Exception as e:
            self.failure_count += 1
            self.last_failure_time = time.time()
            
            if self.failure_count >= self.failure_threshold:
                self.state = "OPEN"
                logger.error(f"🚨 Circuit breaker OPEN after {self.failure_count} failures")
            
            raise

# Global circuit breaker for Harbor API calls
harbor_circuit_breaker = CircuitBreaker(failure_threshold=10, recovery_timeout=120.0) 