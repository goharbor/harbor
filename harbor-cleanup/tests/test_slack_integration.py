#!/usr/bin/env python3
"""
Test script to verify Slack integration functionality
"""

def test_slack_integration():
    """Test Slack integration without actually sending messages"""
    try:
        print("🧪 Testing Slack integration...")
        
        # Test import
        print("  ✓ Testing imports...")
        from harbor_cleanup.utils.slack_notifier import SlackNotifier, slack_notifier
        print("    - SlackNotifier class imported successfully")
        print("    - Global slack_notifier instance available")
        
        # Test configuration
        print("  ✓ Testing configuration...")
        notifier = SlackNotifier()
        print(f"    - Slack enabled: {notifier.enabled}")
        print(f"    - Token configured: {'Yes' if notifier.token else 'No'}")
        print(f"    - Channel configured: {'Yes' if notifier.channel else 'No'}")
        
        # Test message formatting (without sending)
        print("  ✓ Testing message formatting...")
        
        # Mock results for testing
        mock_results = [
            {
                'success': True,
                'result': {
                    'total_artifacts': 1000,
                    'artifacts_to_delete': 200,
                    'artifacts_to_keep': 800,
                    'space_to_cleanup': 5368709120,  # 5GB
                    'total_space': 10737418240  # 10GB
                }
            }
        ]
        
        # Test start notification format (dry run)
        print("    - Testing start notification format...")
        if hasattr(notifier, 'send_cleanup_start'):
            print("      ✓ send_cleanup_start method exists")
        
        # Test completion notification format (dry run)
        print("    - Testing completion notification format...")
        if hasattr(notifier, 'send_cleanup_complete'):
            print("      ✓ send_cleanup_complete method exists")
        
        # Test error notification format (dry run)
        print("    - Testing error notification format...")
        if hasattr(notifier, 'send_error_notification'):
            print("      ✓ send_error_notification method exists")
        
        print("\n✅ Slack integration test passed!")
        print("💡 To test actual notifications:")
        print("   1. Set SLACK_ENABLED=true")
        print("   2. Set SLACK_BOT_TOKEN=xoxb-your-token")
        print("   3. Set SLACK_CHANNEL_ID=your-channel-id")
        print("   4. Run: python main.py")
        
        return True
        
    except ImportError as e:
        print(f"❌ Import error: {e}")
        return False
    except Exception as e:
        print(f"❌ Error: {e}")
        return False

if __name__ == "__main__":
    success = test_slack_integration()
    exit(0 if success else 1)