#!/usr/bin/env python3
"""
Test script to verify the package structure is correct
"""

def test_package_structure():
    """Test that the package structure is correct"""
    try:
        print("🧪 Testing package structure...")
        
        # Test core package
        print("  ✓ Testing core package...")
        from harbor_cleanup.core import config
        from harbor_cleanup.core import retention_policy
        print(f"    - Config loaded: {config.PROJECT}")
        print(f"    - Retention policy class: {retention_policy.RetentionPolicy}")
        
        # Test utils package  
        print("  ✓ Testing utils package...")
        from harbor_cleanup.utils import formatting
        print(f"    - Format 1GB: {formatting.format_size(1024**3)}")
        print(f"    - Format 1h: {formatting.format_duration(3600)}")
        
        # Test main package imports
        print("  ✓ Testing main package...")
        import harbor_cleanup
        print(f"    - Package version: {harbor_cleanup.__version__}")
        print(f"    - Package author: {harbor_cleanup.__author__}")
        
        print("\n✅ Package structure test passed!")
        print("📁 Folder organization:")
        print("   ├── harbor_cleanup/")
        print("   │   ├── core/        (config, processor, retention_policy)")
        print("   │   ├── api/         (harbor_api)")
        print("   │   └── utils/       (formatting, reporting)")
        print("   ├── tests/           (test modules)")
        print("   ├── docs/            (documentation)")
        print("   └── main.py          (entry point)")
        
        return True
        
    except ImportError as e:
        print(f"❌ Import error: {e}")
        return False
    except Exception as e:
        print(f"❌ Error: {e}")
        return False

if __name__ == "__main__":
    success = test_package_structure()
    exit(0 if success else 1) 