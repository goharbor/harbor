from __future__ import absolute_import

import unittest
import swagger_client
import time

from testutils import ADMIN_CLIENT
from library.user import User
from library.system import System


class TestSysCVEWhitelist(unittest.TestCase):
    """
    Test case:
        System Level CVE Whitelist
    Setup:
        Create user(RA)
    Test Steps:
        1. User(RA) reads the system level CVE whitelist and it's empty.
        2. User(RA) updates the system level CVE whitelist, verify it's failed.
        3. Update user(RA) to system admin
        4. User(RA) updates the system level CVE whitelist, verify it's successful.
        5. User(RA) reads the system level CVE whitelist, verify the CVE list is updated.
        6. User(RA) updates the expiration date of system level CVE whitelist.
        7. User(RA) reads the system level CVE whitelist, verify the expiration date is updated.
    Tear Down:
        1. Clear the system level CVE whitelist.
        2. Delete User(RA)
    """
    def setUp(self):
        self.user = User()
        self.system = System()
        user_ra_password = "Aa123456"
        print("Setup: Creating user for test")
        user_ra_id, user_ra_name = self.user.create_user(user_password=user_ra_password, **ADMIN_CLIENT)
        print("Created user: %s, id: %s" % (user_ra_name, user_ra_id))
        self.USER_RA_CLIENT = dict(endpoint=ADMIN_CLIENT["endpoint"],
                                   username=user_ra_name,
                                   password=user_ra_password)
        self.user_ra_id = int(user_ra_id)

    def testSysCVEWhitelist(self):
        # 1. User(RA) reads the system level CVE whitelist and it's empty.
        wl = self.system.get_cve_whitelist(**self.USER_RA_CLIENT)
        self.assertEqual(0, len(wl.items), "The initial system level CVE whitelist is not empty: %s" % wl.items)
        # 2. User(RA) updates the system level CVE whitelist, verify it's failed.
        cves = ['CVE-2019-12310']
        self.system.set_cve_whitelist(None, 403, *cves, **self.USER_RA_CLIENT)
        # 3. Update user(RA) to system admin
        self.user.update_user_role_as_sysadmin(self.user_ra_id, True, **ADMIN_CLIENT)
        # 4. User(RA) updates the system level CVE whitelist, verify it's successful.
        self.system.set_cve_whitelist(None, 200, *cves, **self.USER_RA_CLIENT)
        # 5. User(RA) reads the system level CVE whitelist, verify the CVE list is updated.
        expect_wl = [swagger_client.CVEWhitelistItem(cve_id='CVE-2019-12310')]
        wl = self.system.get_cve_whitelist(**self.USER_RA_CLIENT)
        self.assertIsNone(wl.expires_at)
        self.assertEqual(expect_wl, wl.items)
        # 6. User(RA) updates the expiration date of system level CVE whitelist.
        exp = int(time.time()) + 3600
        self.system.set_cve_whitelist(exp, 200, *cves, **self.USER_RA_CLIENT)
        # 7. User(RA) reads the system level CVE whitelist, verify the expiration date is updated.
        wl = self.system.get_cve_whitelist(**self.USER_RA_CLIENT)
        self.assertEqual(exp, wl.expires_at)

    def tearDown(self):
        print("TearDown: Clearing the Whitelist")
        self.system.set_cve_whitelist(**ADMIN_CLIENT)
        print("TearDown: Deleting user: %d" % self.user_ra_id)
        self.user.delete_user(self.user_ra_id, **ADMIN_CLIENT)


if __name__ == '__main__':
    unittest.main()