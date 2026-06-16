#!/usr/bin/env python3
"""
Diagnostic script to test PAT API behavior
Run this to identify actual issues with PAT list endpoint
"""
import sys
import os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from testutils import ADMIN_CLIENT
from library.user import User

def test_pat_list_behavior():
    """
    Diagnostic test to verify PAT list API returns tokens when available
    """
    user = User()
    url = ADMIN_CLIENT["endpoint"]
    user_password = "Diagnostic123456"

    print("\n" + "="*70)
    print("PAT LIST API DIAGNOSTIC TEST")
    print("="*70)

    try:
        # Step 1: Create test user
        print("\n[1/6] Creating test user...")
        user_id, user_name = user.create_user(
            user_password=user_password,
            **ADMIN_CLIENT
        )
        print(f"✓ User created: ID={user_id}, Name={user_name}")

        USER_CLIENT = dict(
            endpoint=url,
            username=user_name,
            password=user_password
        )

        # Step 2: List empty PATs
        print("\n[2/6] Listing PATs before creation (should be empty)...")
        pats_empty = user.list_personal_access_tokens(user_id, **USER_CLIENT)
        print(f"✓ Empty list response: {len(pats_empty)} tokens")
        if pats_empty:
            print("  WARNING: Empty list returned non-empty array!")
            for pat in pats_empty:
                print(f"    - {pat.name} (ID: {pat.id})")

        # Step 3: Create first PAT
        print("\n[3/6] Creating first PAT...")
        pat1_data, pat1_id = user.create_personal_access_token(
            user_id,
            name="diagnostic-pat-1",
            description="First diagnostic token",
            **USER_CLIENT
        )
        print(f"✓ PAT created: ID={pat1_id}")
        print(f"  - Has secret: {bool(pat1_data.secret)}")
        print(f"  - Secret format: {pat1_data.secret[:20]}..." if pat1_data.secret else "  - No secret")

        # Step 4: List with 1 token
        print("\n[4/6] Listing PATs after creating 1 token...")
        pats_one = user.list_personal_access_tokens(user_id, **USER_CLIENT)
        print(f"✓ List response: {len(pats_one)} token(s)")
        if len(pats_one) > 0:
            for pat in pats_one:
                print(f"  ✓ {pat.name} (ID: {pat.id}, Disabled: {pat.disabled})")
        else:
            print("  ✗ ISSUE: Created token not returned in list!")

        # Step 5: Create second PAT
        print("\n[5/6] Creating second PAT...")
        pat2_data, pat2_id = user.create_personal_access_token(
            user_id,
            name="diagnostic-pat-2",
            description="Second diagnostic token",
            **USER_CLIENT
        )
        print(f"✓ PAT created: ID={pat2_id}")

        # Step 6: List with 2 tokens
        print("\n[6/6] Listing PATs after creating 2 tokens...")
        pats_two = user.list_personal_access_tokens(user_id, **USER_CLIENT)
        print(f"✓ List response: {len(pats_two)} token(s)")
        if len(pats_two) > 0:
            for pat in pats_two:
                print(f"  ✓ {pat.name} (ID: {pat.id}, Disabled: {pat.disabled})")
        else:
            print("  ✗ ISSUE: Created tokens not returned in list!")

        # Cleanup
        print("\n[Cleanup] Deleting test user...")
        user.delete_user(user_id, **ADMIN_CLIENT)
        print("✓ User deleted")

        # Analysis
        print("\n" + "="*70)
        print("DIAGNOSTIC RESULTS")
        print("="*70)
        print(f"Empty list returned: {len(pats_empty)} items (expected 0)")
        print(f"After 1 creation: {len(pats_one)} items (expected 1)")
        print(f"After 2 creations: {len(pats_two)} items (expected 2)")

        if len(pats_empty) == 0 and len(pats_one) == 1 and len(pats_two) == 2:
            print("\n✓✓✓ API WORKING CORRECTLY - List returns tokens as expected!")
            return True
        else:
            print("\n✗✗✗ API ISSUE DETECTED:")
            if len(pats_one) != 1:
                print("  - Tokens are not being returned in list after creation")
                print("  - Possible causes:")
                print("    1. Query filtering by user_id is not working")
                print("    2. Tokens are not being persisted to database")
                print("    3. Pagination defaults are limiting results")
            return False

    except Exception as e:
        print(f"\n✗ ERROR: {e}")
        import traceback
        traceback.print_exc()
        return False

if __name__ == '__main__':
    success = test_pat_list_behavior()
    sys.exit(0 if success else 1)
