import textwrap

import pytest
import yaml

from utils.migration import read_conf
from migrations.version_2_15_0 import migrate as migrate_2_15_0
from migrations.version_2_16_0 import migrate as migrate_2_16_0


def _write_yml(path, body):
    path.write_text(textwrap.dedent(body).lstrip())
    return str(path)


def test_read_conf_missing_strong_ssl_ciphers_defaults_false(tmp_path):
    cfg = _write_yml(
        tmp_path / "harbor.yml",
        """
        _version: 2.14.0
        hostname: harbor.example
        https:
          port: 443
          certificate: /path/cert.pem
          private_key: /path/key.pem
        """,
    )
    assert read_conf(cfg)["strong_ssl_ciphers"] is False


@pytest.mark.parametrize("value", [True, False])
def test_read_conf_preserves_https_strong_ssl_ciphers(tmp_path, value):
    cfg = _write_yml(
        tmp_path / "harbor.yml",
        f"""
        _version: 2.14.0
        hostname: harbor.example
        https:
          port: 443
          certificate: /path/cert.pem
          private_key: /path/key.pem
          strong_ssl_ciphers: {str(value).lower()}
        """,
    )
    assert read_conf(cfg)["strong_ssl_ciphers"] is value


@pytest.mark.parametrize("value", [True, False])
def test_read_conf_preserves_internal_tls_strong_ssl_ciphers(tmp_path, value):
    cfg = _write_yml(
        tmp_path / "harbor.yml",
        f"""
        _version: 2.14.0
        hostname: harbor.example
        https:
          port: 443
          certificate: /path/cert.pem
          private_key: /path/key.pem
        internal_tls:
          enabled: true
          dir: /etc/harbor/tls/internal
          strong_ssl_ciphers: {str(value).lower()}
        """,
    )
    assert read_conf(cfg)["strong_ssl_ciphers"] is value


def test_read_conf_ors_https_false_with_internal_true(tmp_path):
    cfg = _write_yml(
        tmp_path / "harbor.yml",
        """
        _version: 2.14.0
        hostname: harbor.example
        https:
          port: 443
          certificate: /path/cert.pem
          private_key: /path/key.pem
          strong_ssl_ciphers: false
        internal_tls:
          enabled: true
          dir: /etc/harbor/tls/internal
          strong_ssl_ciphers: true
        """,
    )
    assert read_conf(cfg)["strong_ssl_ciphers"] is True


def _assert_migrate_writes_false(migrate_fn, input_version, tmp_path):
    input_cfg = _write_yml(
        tmp_path / "in.yml",
        f"""
        _version: {input_version}
        hostname: harbor.example
        https:
          port: 443
          certificate: /path/cert.pem
          private_key: /path/key.pem
        """,
    )
    output_cfg = str(tmp_path / "out.yml")
    migrate_fn(input_cfg, output_cfg)

    rendered = open(output_cfg).read()
    out = yaml.safe_load(rendered)
    assert out["https"]["strong_ssl_ciphers"] is False
    assert "strong_ssl_ciphers: none" not in rendered
    assert "strong_ssl_ciphers: false" in rendered


def test_migrate_2_15_0_missing_strong_ssl_ciphers_writes_false(tmp_path):
    _assert_migrate_writes_false(migrate_2_15_0, "2.14.0", tmp_path)


def test_migrate_2_16_0_missing_strong_ssl_ciphers_writes_false(tmp_path):
    _assert_migrate_writes_false(migrate_2_16_0, "2.15.0", tmp_path)
