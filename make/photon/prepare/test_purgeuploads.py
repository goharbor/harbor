
import unittest
from models import PurgeUpload

class TestPurgeUploadsDefault(unittest.TestCase):
    def test_validate_config(self):
        purge_config = dict([
            ('enabled',True),
            ('age','168h'),
            ('interval','24h'),
            ('dryrun', False),
        ])
        cfg = PurgeUpload(purge_config)
        cfg.validate()

    def test_validate_config(self):
        purge_config = dict([
            ('enabled','false'),
            ('age','168h'),
            ('interval','24h'),
            ('dryrun', 'false'),
        ])
        cfg = PurgeUpload(purge_config)
        cfg.validate()

    def test_validate_config_2hour(self):
        purge_config = dict([
            ('enabled',True),
            ('age','2h'),
            ('interval','2h'),
            ('dryrun', False),
        ])
        cfg = PurgeUpload(purge_config)
        cfg.validate()        

    def test_validate_config_1hour(self):
        purge_config = dict([
            ('enabled',True),
            ('age','1h'),
            ('interval','1h'),
            ('dryrun', False),
        ])
        cfg = PurgeUpload(purge_config)
        with self.assertRaises(Exception):
            cfg.validate()

    def test_validate_config_invalid_format(self):
        purge_config = dict([
            ('enabled',True),
            ('age','1s'),
            ('interval','1s'),
            ('dryrun', False),
        ])
        cfg = PurgeUpload(purge_config)
        with self.assertRaises(Exception):
            cfg.validate()      

    def test_validate_config_invalid_format(self):
        purge_config = dict([
            ('enabled',True),
            ('age',168),
            ('interval',24),
            ('dryrun', False),
        ])
        cfg = PurgeUpload(purge_config)
        with self.assertRaises(Exception):
            cfg.validate()  

    def test_validate_config_disabled_invalid_format(self):
        purge_config = dict([
            ('enabled',"false"),
            ('age','ssh'),
            ('interval','ssh'),
            ('dryrun', False),
        ])
        cfg = PurgeUpload(purge_config)
        cfg.validate()  

    def test_validate_config_invalid_string(self):
        purge_config = dict([
            ('enabled',True),
            ('age','ssh'),
            ('interval','ssh'),
            ('dryrun', False),
        ])
        cfg = PurgeUpload(purge_config)
        with self.assertRaises(Exception):
            cfg.validate()
