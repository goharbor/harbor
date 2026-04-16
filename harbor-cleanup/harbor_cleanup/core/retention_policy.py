"""
Harbor Cleanup Tool - Retention Policy Module

This module handles artifact categorization and retention policy logic.
"""

import os
from datetime import datetime
from .config import VERBOSE_ARTIFACTS, logger


class RetentionPolicy:
    """Centralized retention policy configuration"""
    
    def __init__(self):
        # Age-based policies (days)
        self.age_policies = {
            'prod': int(os.getenv("PROD_LABEL_CLEANUP_DAYS", "45")),
            'non-prod': int(os.getenv("NON_PROD_LABEL_CLEANUP_DAYS", "14")),
            'no-label': int(os.getenv("EMPTY_LABEL_CLEANUP_DAYS", "45")),
            'other': int(os.getenv("OTHER_LABEL_CLEANUP_DAYS", os.getenv("EMPTY_LABEL_CLEANUP_DAYS", "45")))
        }
        
        # Count-based policies (number of artifacts to retain regardless of age)
        self.count_policies = {
            'prod': int(os.getenv("RETAIN_LAST_PROD", "100")),
            'non-prod': int(os.getenv("RETAIN_LAST_NON_PROD", "0")),
            'no-label': int(os.getenv("RETAIN_LAST_NO_LABEL", "500")),
            'other': int(os.getenv("RETAIN_LAST_OTHER", "0"))
        }
        
        # Label matching configuration
        base_mappings = {
            'prod': ['prod', 'production', 'release', 'stable', 'main', 'master'],
            'non-prod': ['nonprod', 'non-prod', 'staging', 'stage', 'dev', 'development', 'test', 'testing', 'qa', 'uat', 'beta', 'alpha', 'feature', 'develop'],
            'no-label': [],  # Special case for artifacts without labels
            'other': []      # Catch-all for other labels
        }
        
        # Add additional labels from environment variables
        additional_prod = os.getenv("ADDITIONAL_PROD_LABELS", "")
        additional_nonprod = os.getenv("ADDITIONAL_NONPROD_LABELS", "")
        
        if additional_prod:
            base_mappings['prod'].extend([label.strip().lower() for label in additional_prod.split(',') if label.strip()])
        
        if additional_nonprod:
            base_mappings['non-prod'].extend([label.strip().lower() for label in additional_nonprod.split(',') if label.strip()])
        
        self.label_mappings = base_mappings
    
    def get_age_policy(self, category):
        """Get age-based retention policy for a category"""
        return self.age_policies.get(category, self.age_policies['other'])
    
    def get_count_policy(self, category):
        """Get count-based retention policy for a category"""
        return self.count_policies.get(category, self.count_policies['other'])
    
    def categorize_artifact(self, labels):
        """Categorize an artifact based on its labels"""
        if not labels:
            return 'no-label'
        
        label_names = [label.get("name", "").lower() for label in labels]
        
        # Debug logging for label categorization
        if VERBOSE_ARTIFACTS:
            logger.debug(f"🏷️  Label names found: {label_names}")
            logger.debug(f"🏷️  Prod patterns: {self.label_mappings['prod']}")
            logger.debug(f"🏷️  Non-prod patterns: {self.label_mappings['non-prod']}")
        
        # Check for exact matches only
        for label_name in label_names:
            if label_name in self.label_mappings['non-prod']:
                if VERBOSE_ARTIFACTS:
                    logger.debug(f"🏷️  Exact NON-PROD match: '{label_name}'")
                return 'non-prod'
            if label_name in self.label_mappings['prod']:
                if VERBOSE_ARTIFACTS:
                    logger.debug(f"🏷️  Exact PROD match: '{label_name}'")
                return 'prod'
        
        # Everything else is 'other'
        if VERBOSE_ARTIFACTS:
            logger.debug(f"🏷️  No matches found, categorizing as OTHER: {label_names}")
        return 'other'
    
    def should_delete_by_age(self, artifact, category):
        """Determine if artifact should be deleted based on age policy"""
        # Get timestamps
        pull_time_str = artifact.get("pull_time")
        push_time_str = artifact.get("push_time")
        
        # Use pull_time primarily, fall back to push_time
        time_to_use = None
        if pull_time_str and pull_time_str != "0001-01-01T00:00:00.000Z":
            time_to_use = pull_time_str
        elif push_time_str and push_time_str != "0001-01-01T00:00:00.000Z":
            time_to_use = push_time_str
        
        if not time_to_use:
            return False
        
        try:
            # Parse the timestamp
            time_to_use = time_to_use.split('.')[0] + 'Z'
            artifact_time = datetime.fromisoformat(time_to_use.replace('Z', '+00:00'))
            
            # Calculate age
            now = datetime.now(artifact_time.tzinfo)
            age_days = (now - artifact_time).days
            
            # Check against policy
            policy_days = self.get_age_policy(category)
            return age_days > policy_days
            
        except ValueError:
            return False
    
    def get_summary(self):
        """Get a summary of all policies for display"""
        summary = []
        summary.append("Retention Policies:")
        summary.append("Age-based policies:")
        for category, days in self.age_policies.items():
            summary.append(f"  - {category}: {days} days")
        
        summary.append("Count-based policies:")
        for category, count in self.count_policies.items():
            summary.append(f"  - {category}: keep last {count} artifacts")
        
        summary.append("Label mappings:")
        for category, labels in self.label_mappings.items():
            if labels:
                summary.append(f"  - {category}: {', '.join(labels)}")
            elif category == 'no-label':
                summary.append(f"  - {category}: (artifacts without labels)")
            elif category == 'other':
                summary.append(f"  - {category}: (catch-all for unmatched labels)")
        
        return "\n".join(summary) 