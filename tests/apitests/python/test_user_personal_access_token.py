from __future__ import absolute_import

import unittest
import time
from testutils import ADMIN_CLIENT, suppress_urllib3_warning
from library.user import User


class TestUserPersonalAccessToken(unittest.TestCase):

    @suppress_urllib3_warning
    def setUp(self):
        self.user = User()
        self.url = ADMIN_CLIENT["endpoint"]
        self.user_password = "Aa123456"

    def tearDown(self):
        pass

    @suppress_urllib3_warning
    def testPersonalAccessTokenCRUD(self):
        """
        Test case:
            Personal Access Token CRUD operations
        Test step and expected result:
            1. Create a test user (TU);
            2. Create a personal access token (PAT) for the test user;
            3. List all PATs for the test user, should have 1 token;
            4. Get the PAT by ID, verify token details;
            5. Update the PAT (disable it);
            6. Verify the PAT is disabled;
            7. Create another PAT;
            8. List PATs, should have 2 tokens;
            9. Delete first PAT;
            10. List PATs, should have 1 token;
            11. Delete the test user;
        """

        # 1. Create a test user (TU)
        user_id, user_name = self.user.create_user(
            user_password=self.user_password,
            **ADMIN_CLIENT
        )
        USER_CLIENT = dict(
            endpoint=self.url,
            username=user_name,
            password=self.user_password
        )

        # 2. Create a personal access token (PAT) for the test user
        pat_name = "test-pat-{}".format(int(round(time.time() * 1000)))
        pat_description = "Test PAT for integration testing"
        pat_data, pat_id = self.user.create_personal_access_token(
            user_id,
            name=pat_name,
            description=pat_description,
            expires_in_days=30,
            **ADMIN_CLIENT
        )
        self.assertIsNotNone(pat_data)
        self.assertIsNotNone(pat_data.secret)
        self.assertTrue(pat_data.secret.startswith("hbr_pat_"))

        # 3. List all PATs for the test user, should have 1 token
        pats = self.user.list_personal_access_tokens(user_id, **USER_CLIENT)
        self.assertIsNotNone(pats)
        self.assertEqual(len(pats), 1)

        # 4. Get the PAT by ID, verify token details
        pat = self.user.get_personal_access_token(user_id, pat_id, **USER_CLIENT)
        self.assertIsNotNone(pat)
        self.assertEqual(pat.id, pat_id)
        self.assertEqual(pat.name, pat_name)
        self.assertEqual(pat.description, pat_description)
        self.assertEqual(pat.user_id, user_id)
        self.assertFalse(pat.disabled)
        self.assertIsNone(pat.secret)  # Secret should not be exposed in GET

        # 5. Update the PAT (disable it)
        self.user.update_personal_access_token(
            user_id,
            pat_id,
            disabled=True,
            **ADMIN_CLIENT
        )

        # 6. Verify the PAT is disabled
        pat = self.user.get_personal_access_token(user_id, pat_id, **USER_CLIENT)
        self.assertTrue(pat.disabled)

        # 7. Create another PAT
        pat_name_2 = "test-pat-2-{}".format(int(round(time.time() * 1000)))
        pat_data_2, pat_id_2 = self.user.create_personal_access_token(
            user_id,
            name=pat_name_2,
            description="Second test PAT",
            **USER_CLIENT
        )
        self.assertIsNotNone(pat_data_2)
        self.assertNotEqual(pat_id, pat_id_2)

        # 8. List PATs, should have 2 tokens
        pats = self.user.list_personal_access_tokens(user_id, **USER_CLIENT)
        self.assertEqual(len(pats), 2)

        # Verify both PATs are in the list
        pat_ids_in_list = [p.id for p in pats]
        self.assertIn(pat_id, pat_ids_in_list)
        self.assertIn(pat_id_2, pat_ids_in_list)

        # 9. Delete first PAT
        self.user.delete_personal_access_token(user_id, pat_id, **USER_CLIENT)

        # 10. List PATs, should have 1 token
        pats = self.user.list_personal_access_tokens(user_id, **USER_CLIENT)
        self.assertEqual(len(pats), 1)
        self.assertEqual(pats[0].id, pat_id_2)

        # 11. Delete the test user
        self.user.delete_user(user_id, **ADMIN_CLIENT)
        self.user.get_user_by_id(
            user_id,
            expect_status_code=404,
            **ADMIN_CLIENT
        )

    @suppress_urllib3_warning
    def testPersonalAccessTokenPagination(self):
        """
        Test case:
            Personal Access Token pagination
        Test step and expected result:
            1. Create a test user (TU);
            2. Create 15 PATs for the test user;
            3. List PATs with page=1, page_size=10, should have 10 tokens;
            4. List PATs with page=2, page_size=10, should have 5 tokens;
            5. List PATs with page_size=20, should have 15 tokens on page 1;
            6. Delete the test user;
        """

        # 1. Create a test user (TU)
        user_id, user_name = self.user.create_user(
            user_password=self.user_password,
            **ADMIN_CLIENT
        )
        USER_CLIENT = dict(
            endpoint=self.url,
            username=user_name,
            password=self.user_password
        )

        # 2. Create 15 PATs for the test user
        timestamp = int(round(time.time() * 1000))
        for i in range(1, 16):
            pat_name = "paginated-pat-{}-{}".format(i, timestamp)
            self.user.create_personal_access_token(
                user_id,
                name=pat_name,
                **ADMIN_CLIENT
            )

        # 3. List PATs with page=1, page_size=10, should have 10 tokens
        pats_page1 = self.user.list_personal_access_tokens(
            user_id,
            page=1,
            page_size=10,
            **USER_CLIENT
        )
        self.assertEqual(len(pats_page1), 10)

        # 4. List PATs with page=2, page_size=10, should have 5 tokens
        pats_page2 = self.user.list_personal_access_tokens(
            user_id,
            page=2,
            page_size=10,
            **USER_CLIENT
        )
        self.assertEqual(len(pats_page2), 5)

        # Verify no overlap between pages
        page1_ids = [p.id for p in pats_page1]
        page2_ids = [p.id for p in pats_page2]
        self.assertEqual(len(set(page1_ids) & set(page2_ids)), 0)

        # 5. List PATs with page_size=20, should have 15 tokens on page 1
        pats_large_page = self.user.list_personal_access_tokens(
            user_id,
            page=1,
            page_size=20,
            **USER_CLIENT
        )
        self.assertEqual(len(pats_large_page), 15)

        # 6. Delete the test user
        self.user.delete_user(user_id, **ADMIN_CLIENT)

    @suppress_urllib3_warning
    def testPersonalAccessTokenExpiry(self):
        """
        Test case:
            Personal Access Token expiry handling
        Test step and expected result:
            1. Create a test user (TU);
            2. Create a PAT that never expires (expires_in_days=0 or -1);
            3. Create a PAT with 30 day expiry;
            4. Create a PAT with 1 day expiry;
            5. Verify expiry times are set correctly;
            6. Delete the test user;
        """

        # 1. Create a test user (TU)
        user_id, user_name = self.user.create_user(
            user_password=self.user_password,
            **ADMIN_CLIENT
        )
        USER_CLIENT = dict(
            endpoint=self.url,
            username=user_name,
            password=self.user_password
        )

        # 2. Create a PAT that never expires
        pat_never_expire, pat_id_never = self.user.create_personal_access_token(
            user_id,
            name="never-expire-pat",
            expires_in_days=0,
            **USER_CLIENT
        )
        self.assertIsNotNone(pat_never_expire)

        # 3. Create a PAT with 30 day expiry
        pat_30day, pat_id_30 = self.user.create_personal_access_token(
            user_id,
            name="30day-expiry-pat",
            expires_in_days=30,
            **USER_CLIENT
        )
        self.assertIsNotNone(pat_30day)

        # 4. Create a PAT with 1 day expiry
        pat_1day, pat_id_1 = self.user.create_personal_access_token(
            user_id,
            name="1day-expiry-pat",
            expires_in_days=1,
            **USER_CLIENT
        )
        self.assertIsNotNone(pat_1day)

        # 5. Verify expiry times are set correctly
        # Get tokens to verify expiry settings
        current_time = int(round(time.time()))

        # Never expire token
        pat_never = self.user.get_personal_access_token(
            user_id,
            pat_id_never,
            **USER_CLIENT
        )
        # Verify it has a far future expiry or special "never" value
        self.assertTrue(pat_never.expires_at == -1 or pat_never.expires_at > current_time + 86400 * 29)

        # 30 day expiry token
        pat_30 = self.user.get_personal_access_token(
            user_id,
            pat_id_30,
            **USER_CLIENT
        )
        # Verify it expires in ~30 days (within 1 hour margin)
        expected_30day = current_time + 86400 * 30
        self.assertTrue(abs(pat_30.expires_at - expected_30day) < 3600)

        # 1 day expiry token
        pat_1 = self.user.get_personal_access_token(
            user_id,
            pat_id_1,
            **USER_CLIENT
        )
        # Verify it expires in ~1 day (within 1 hour margin)
        expected_1day = current_time + 86400
        self.assertTrue(abs(pat_1.expires_at - expected_1day) < 3600)

        # 6. Delete the test user
        self.user.delete_user(user_id, **ADMIN_CLIENT)

    @suppress_urllib3_warning
    def testPersonalAccessTokenAuthorizationOwnTokens(self):
        """
        Test case:
            User can list and manage their own PATs
        Test step and expected result:
            1. Create a test user (TU);
            2. TU creates a PAT as themselves;
            3. TU lists their own PATs, should see the token;
            4. TU gets their own PAT by ID;
            5. TU disables their own PAT;
            6. TU deletes their own PAT;
            7. Delete the test user;
        """

        # 1. Create a test user (TU)
        user_id, user_name = self.user.create_user(
            user_password=self.user_password,
            **ADMIN_CLIENT
        )
        USER_CLIENT = dict(
            endpoint=self.url,
            username=user_name,
            password=self.user_password
        )

        # 2. TU creates a PAT as themselves
        pat_data, pat_id = self.user.create_personal_access_token(
            user_id,
            name="self-created-pat",
            **USER_CLIENT
        )
        self.assertIsNotNone(pat_data)

        # 3. TU lists their own PATs, should see the token
        pats = self.user.list_personal_access_tokens(user_id, **USER_CLIENT)
        self.assertEqual(len(pats), 1)
        self.assertEqual(pats[0].id, pat_id)

        # 4. TU gets their own PAT by ID
        pat = self.user.get_personal_access_token(user_id, pat_id, **USER_CLIENT)
        self.assertIsNotNone(pat)
        self.assertEqual(pat.id, pat_id)

        # 5. TU disables their own PAT
        self.user.update_personal_access_token(
            user_id,
            pat_id,
            disabled=True,
            **USER_CLIENT
        )
        pat = self.user.get_personal_access_token(user_id, pat_id, **USER_CLIENT)
        self.assertTrue(pat.disabled)

        # 6. TU deletes their own PAT
        self.user.delete_personal_access_token(user_id, pat_id, **USER_CLIENT)
        pats = self.user.list_personal_access_tokens(user_id, **USER_CLIENT)
        self.assertEqual(len(pats), 0)

        # 7. Delete the test user
        self.user.delete_user(user_id, **ADMIN_CLIENT)


if __name__ == '__main__':
    unittest.main()
