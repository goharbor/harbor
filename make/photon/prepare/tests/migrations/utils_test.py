import pytest
import importlib

from migration.utils import search

class mockModule:
    def __init__(self, revision, down_revision):
        self.revision = revision
        self.down_revision = down_revision

def mock_import_module_loop(module_path: str):
    loop_modules = {
    'migration.versions.1_9_0': mockModule('1.9.0', None),
    'migration.versions.1_10_0': mockModule('1.10.0', '2.0.0'),
    'migration.versions.2_0_0': mockModule('2.0.0', '1.10.0')
    }
    return loop_modules[module_path]

def mock_import_module_mission(module_path: str):
    loop_modules = {
    'migration.versions.1_9_0': mockModule('1.9.0', None),
    'migration.versions.1_10_0': mockModule('1.10.0', None),
    'migration.versions.2_0_0': mockModule('2.0.0', '1.10.0')
    }
    return loop_modules[module_path]

def mock_import_module_success(module_path: str):
    loop_modules = {
    'migration.versions.1_9_0': mockModule('1.9.0', None),
    'migration.versions.1_10_0': mockModule('1.10.0', '1.9.0'),
    'migration.versions.2_0_0': mockModule('2.0.0', '1.10.0')
    }
    return loop_modules[module_path]

@pytest.fixture
def mock_import_module_with_loop(monkeypatch):
    monkeypatch.setattr(importlib, "import_module", mock_import_module_loop)

@pytest.fixture
def mock_import_module_with_mission(monkeypatch):
    monkeypatch.setattr(importlib, "import_module", mock_import_module_mission)

@pytest.fixture
def mock_import_module_with_success(monkeypatch):
    monkeypatch.setattr(importlib, "import_module", mock_import_module_success)

def test_search_loop(mock_import_module_with_loop):
    with pytest.raises(Exception):
        search('1.9.0', '2.0.0')

def test_search_mission(mock_import_module_with_mission):
    with pytest.raises(Exception):
        search('1.9.0', '2.0.0')

def test_search_success():
    migration_path = search('1.9.0', '2.0.0')
    assert migration_path[0].revision == '1.10.0'
    assert migration_path[1].revision == '2.0.0'
