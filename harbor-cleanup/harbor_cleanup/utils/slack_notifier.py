"""
Harbor Cleanup Tool - Slack Notifier Module

This module handles Slack notifications for cleanup start/end summaries.
"""

import json
import requests
from datetime import datetime
from zoneinfo import ZoneInfo
from ..core.config import (
    SLACK_ENABLED, SLACK_BOT_TOKEN, SLACK_CHANNEL_ID,
    PROJECT, DRY_RUN, TRIGGER_GC_AFTER_CLEANUP, HARBOR_ENV, logger
)
from .formatting import format_size, format_duration


class SlackNotifier:
    """Handles Slack notifications for Harbor cleanup operations"""
    
    def __init__(self):
        self.enabled = SLACK_ENABLED
        self.token = SLACK_BOT_TOKEN
        self.channel = SLACK_CHANNEL_ID
        self.base_url = "https://slack.com/api"
        
        if self.enabled and not self.token:
            logger.warning("⚠️  Slack notifications enabled but SLACK_BOT_TOKEN not provided")
            self.enabled = False
        
        if self.enabled and not self.channel:
            logger.warning("⚠️  Slack notifications enabled but SLACK_CHANNEL_ID not provided")
            self.enabled = False
    
    def _get_ist_timestamp(self):
        """Get current timestamp in IST format"""
        try:
            ist_tz = ZoneInfo("Asia/Kolkata")
            current_time = datetime.now(ist_tz)
            return current_time.strftime("%Y-%m-%d %H:%M:%S IST")
        except Exception as e:
            logger.warning(f"⚠️  Failed to get IST timestamp: {e}, falling back to UTC")
            return datetime.now().strftime("%Y-%m-%d %H:%M:%S UTC")
    
    def _send_message(self, blocks, text_fallback):
        """Send a message to Slack using the Web API"""
        if not self.enabled:
            return False
        
        try:
            headers = {
                'Authorization': f'Bearer {self.token}',
                'Content-Type': 'application/json'
            }
            
            payload = {
                'channel': self.channel,
                'blocks': blocks,
                'text': text_fallback  # Fallback for notifications
            }
            
            response = requests.post(
                f"{self.base_url}/chat.postMessage",
                headers=headers,
                json=payload,
                timeout=10
            )
            
            if response.status_code == 200:
                result = response.json()
                if result.get('ok'):
                    logger.info("✅ Slack notification sent successfully")
                    return True
                else:
                    logger.error(f"❌ Slack API error: {result.get('error', 'Unknown error')}")
                    return False
            else:
                logger.error(f"❌ Slack HTTP error: {response.status_code}")
                return False
                
        except requests.exceptions.RequestException as e:
            logger.error(f"❌ Failed to send Slack notification: {e}")
            return False
        except Exception as e:
            logger.error(f"❌ Unexpected error sending Slack notification: {e}")
            return False
    
    def send_cleanup_start(self, selected_repos=None):
        """Send notification when cleanup starts"""
        if not self.enabled:
            return False
        
        timestamp = self._get_ist_timestamp()
        mode = "🛡️ DRY RUN" if DRY_RUN else "🔥 LIVE RUN"
        repo_info = f"Selected repositories: {', '.join(selected_repos)}" if selected_repos else "All repositories"
        gc_status = "✅ Enabled" if TRIGGER_GC_AFTER_CLEANUP else "❌ Disabled"
        
        blocks = [
            {
                "type": "header",
                "text": {
                    "type": "plain_text",
                    "text": f"🚀 Harbor ({HARBOR_ENV}) Cleanup Started - {mode}"
                }
            },
            {
                "type": "section",
                "fields": [
                    {
                        "type": "mrkdwn",
                        "text": f"*Project:* {PROJECT}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*Started:* {timestamp}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*Mode:* {mode}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*Scope:* {repo_info}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*GC After Cleanup:* {gc_status}"
                    }
                ]
            },
            {
                "type": "context",
                "elements": [
                    {
                        "type": "mrkdwn",
                        "text": "🔍 Analyzing repositories and applying retention policies..."
                    }
                ]
            }
        ]
        
        text_fallback = f"Harbor ({HARBOR_ENV}) Cleanup Started ({mode}) - Project: {PROJECT}"
        return self._send_message(blocks, text_fallback)
    
    def send_cleanup_complete(self, results, script_start_time, interrupted=False):
        """Send notification when cleanup completes"""
        if not self.enabled:
            return False
        
        # Calculate summary statistics
        total_repos_processed = len(results)
        successful_repos = len([r for r in results if r.get('success', False)])
        failed_repos = total_repos_processed - successful_repos
        
        total_artifacts = sum(r['result']['total_artifacts'] for r in results if r.get('success', False))
        total_to_delete = sum(r['result']['artifacts_to_delete'] for r in results if r.get('success', False))
        total_to_keep = sum(r['result']['artifacts_to_keep'] for r in results if r.get('success', False))
        total_cleanup_space = sum(r['result']['space_to_cleanup'] for r in results if r.get('success', False))
        total_space = sum(r['result']['total_space'] for r in results if r.get('success', False))
        
        # Calculate timing
        script_end_time = datetime.now()
        total_elapsed_time = script_end_time.timestamp() - script_start_time
        
        # Determine status and emoji
        if interrupted:
            status = "⚠️ INTERRUPTED"
            status_emoji = "⚠️"
        elif failed_repos > 0:
            status = "⚠️ COMPLETED WITH ISSUES"
            status_emoji = "⚠️"
        else:
            status = "✅ COMPLETED SUCCESSFULLY"
            status_emoji = "✅"
        
        mode = "🛡️ DRY RUN" if DRY_RUN else "🔥 LIVE RUN"
        timestamp = self._get_ist_timestamp()
        gc_status = "✅ Enabled" if TRIGGER_GC_AFTER_CLEANUP else "❌ Disabled"
        
        # Calculate cleanup percentages
        artifact_cleanup_percentage = (total_to_delete / total_artifacts * 100) if total_artifacts > 0 else 0
        space_cleanup_percentage = (total_cleanup_space / total_space * 100) if total_space > 0 else 0
        
        blocks = [
            {
                "type": "header",
                "text": {
                    "type": "plain_text",
                    "text": f"{status_emoji} Harbor ({HARBOR_ENV}) Cleanup {status} - {mode}"
                }
            },
            {
                "type": "section",
                "fields": [
                    {
                        "type": "mrkdwn",
                        "text": f"*Project:* {PROJECT}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*Completed:* {timestamp}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*Duration:* {format_duration(total_elapsed_time)}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*Mode:* {mode}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*GC After Cleanup:* {gc_status}"
                    }
                ]
            },
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": "*📊 Repository Summary:*"
                }
            },
            {
                "type": "section",
                "fields": [
                    {
                        "type": "mrkdwn",
                        "text": f"*Repositories:* {total_repos_processed:,} total"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*Successful:* {successful_repos:,}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*Failed:* {failed_repos:,}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": "*Status:* " + ("All OK" if failed_repos == 0 else f"{failed_repos} issues")
                    }
                ]
            },
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": "*📦 Artifact Summary:*"
                }
            },
            {
                "type": "section",
                "fields": [
                    {
                        "type": "mrkdwn",
                        "text": f"*Total Processed:* {total_artifacts:,}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*To Delete:* {total_to_delete:,}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*To Keep:* {total_to_keep:,}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*Cleanup %:* {artifact_cleanup_percentage:.1f}%"
                    }
                ]
            },
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": "*💾 Storage Summary:*"
                }
            },
            {
                "type": "section",
                "fields": [
                    {
                        "type": "mrkdwn",
                        "text": f"*Total Storage:* {format_size(total_space)}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*To Cleanup:* {format_size(total_cleanup_space)} ({space_cleanup_percentage:.1f}%)"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*Will Remain:* {format_size(total_space - total_cleanup_space)} ({100 - space_cleanup_percentage:.1f}%)"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*Space Saved:* {format_size(total_cleanup_space)} ({space_cleanup_percentage:.1f}%)"
                    }
                ]
            }
        ]
        
        # Add warning for dry run
        if DRY_RUN:
            blocks.append({
                "type": "context",
                "elements": [
                    {
                        "type": "mrkdwn",
                        "text": "⚠️ *This was a DRY RUN* - No artifacts were actually deleted. Run with DRY_RUN=false for live cleanup."
                    }
                ]
            })
        else:
            blocks.append({
                "type": "context",
                "elements": [
                    {
                        "type": "mrkdwn",
                        "text": "🗑️ *LIVE RUN completed* - Artifacts were actually deleted. Garbage collection may take additional time to free up space."
                    }
                ]
            })
        
        text_fallback = f"Harbor ({HARBOR_ENV}) Cleanup {status} ({mode}) - {total_to_delete:,} artifacts deleted, {format_size(total_cleanup_space)} to be freed"
        return self._send_message(blocks, text_fallback)
    
    def send_error_notification(self, error_message):
        """Send notification when cleanup fails with an error"""
        if not self.enabled:
            return False
        
        timestamp = self._get_ist_timestamp()
        mode = "🛡️ DRY RUN" if DRY_RUN else "🔥 LIVE RUN"
        gc_status = "✅ Enabled" if TRIGGER_GC_AFTER_CLEANUP else "❌ Disabled"
        
        blocks = [
            {
                "type": "header",
                "text": {
                    "type": "plain_text",
                    "text": f"❌ Harbor ({HARBOR_ENV}) Cleanup Failed - {mode}"
                }
            },
            {
                "type": "section",
                "fields": [
                    {
                        "type": "mrkdwn",
                        "text": f"*Project:* {PROJECT}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*Failed at:* {timestamp}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*Mode:* {mode}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": "*Status:* Critical Error"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*GC After Cleanup:* {gc_status}"
                    }
                ]
            },
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": f"*Error Details:*\n```{error_message}```"
                }
            },
            {
                "type": "context",
                "elements": [
                    {
                        "type": "mrkdwn",
                        "text": "🔍 Please check the logs for detailed error information and retry if necessary."
                    }
                ]
            }
        ]
        
        text_fallback = f"Harbor ({HARBOR_ENV}) Cleanup Failed ({mode}) - Error: {error_message}"
        return self._send_message(blocks, text_fallback)
    
    def send_gc_notification(self, gc_success, cleanup_size_gb=0):
        """Send notification about garbage collection trigger status"""
        if not self.enabled:
            return False
        
        timestamp = self._get_ist_timestamp()
        mode = "🛡️ DRY RUN" if DRY_RUN else "🔥 LIVE RUN"
        
        # Convert GB to bytes for proper formatting, then format back to human readable
        cleanup_size_bytes = int(cleanup_size_gb * (1024**3))
        cleanup_size_formatted = format_size(cleanup_size_bytes)
        
        if gc_success:
            status_emoji = "🗑️"
            status_text = "Garbage Collection Triggered"
            status_color = "✅"
            status_message = "Successfully triggered Harbor garbage collection"
        else:
            status_emoji = "⚠️"
            status_text = "Garbage Collection Failed"
            status_color = "❌"
            status_message = "Failed to trigger Harbor garbage collection"
        
        blocks = [
            {
                "type": "header",
                "text": {
                    "type": "plain_text",
                    "text": f"{status_emoji} Harbor ({HARBOR_ENV}) {status_text} - {mode}"
                }
            },
            {
                "type": "section",
                "fields": [
                    {
                        "type": "mrkdwn",
                        "text": f"*Project:* {PROJECT}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*Triggered at:* {timestamp}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*Status:* {status_color} {status_message}"
                    },
                    {
                        "type": "mrkdwn",
                        "text": f"*Cleanup Size:* {cleanup_size_formatted}"
                    }
                ]
            }
        ]
        
        # Add context based on success/failure
        if gc_success:
            context_text = "🔄 *Garbage collection is running asynchronously*\n\n"
            context_text += "• Check Harbor UI for GC progress\n"
            context_text += "• Storage space will be freed up during GC process\n"
            
            if cleanup_size_gb > 1000:  # > 1TB
                context_text += "• Large cleanup detected - GC may take several hours\n"
                context_text += "• Monitor Harbor UI and system resources during GC"
            
            if DRY_RUN:
                context_text = "🛡️ *This was a DRY RUN* - GC would be triggered in live mode"
        else:
            context_text = "❌ *GC trigger failed*\n\n"
            context_text += "• Check Harbor connectivity and permissions\n"
            context_text += "• Verify GC is not already running\n"
            context_text += "• Storage space will not be freed automatically\n"
            context_text += "• Consider triggering GC manually from Harbor UI"
        
        blocks.append({
            "type": "context",
            "elements": [
                {
                    "type": "mrkdwn",
                    "text": context_text
                }
            ]
        })
        
        text_fallback = f"Harbor ({HARBOR_ENV}) GC {status_text} ({mode}) - {cleanup_size_formatted} cleanup"
        return self._send_message(blocks, text_fallback)


# Global instance
slack_notifier = SlackNotifier() 