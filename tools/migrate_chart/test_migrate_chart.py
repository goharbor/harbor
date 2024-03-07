
import unittest
import semver
import tempfile
import os
import tarfile
import shutil
from pathlib import Path
from migrate_chart import extract_chart_name_and_version
from migrate_chart import read_chart_version  

class TestExtractChartNameAndVersion(unittest.TestCase):
    def test_valid_chart_name(self):
        filepath = Path("my-project/my-chart-v1.0.0.tgz")
        name, version = extract_chart_name_and_version(filepath)
        self.assertEqual(name, "my-chart")
        self.assertEqual(version, "v1.0.0")

    def test_invalid_chart_name(self):
        filepath = Path("my-project/mychart.tgz")
        name, version = extract_chart_name_and_version(filepath)
        self.assertIsNone(name)
        self.assertIsNone(version)

    # def test_pure_digit(self):
    #     filepath = Path("my-project/my-chart-8.0.0-5.tgz")
    #     name, version = extract_chart_name_and_version(filepath)
    #     self.assertEqual(name, "my-chart")
    #     self.assertEqual(version, "8.0.0")

    # def test_digit_startv(self):
    #     filepath = Path("my-project/my-chart-v8.0.0-5.tgz")
    #     name, version = extract_chart_name_and_version(filepath)
    #     self.assertEqual(name, "my-chart")
    #     self.assertEqual(version, "8.0.0")

    def test_parse_version(self):
        temp_dir = tempfile.mkdtemp()
        file_name = "/Users/daojunz/Downloads/cert-manager/sample/cert-manager-8.0.0-5.tgz"
        name, version = read_chart_version(file_name)
        print(name)
        print(version)

if __name__ == '__main__':
    unittest.main()