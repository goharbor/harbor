import os
import logging
from pathlib import Path
from shutil import copytree, rmtree

from g import internal_tls_dir, DEFAULT_GID, DEFAULT_UID, PG_GID, PG_UID
from utils.misc import check_permission, owner_can_read, get_realpath, port_number_valid
from utils.cert import san_existed


class InternalTLS:

    harbor_certs_filename = {
        'harbor_internal_ca.crt',
        'proxy.crt', 'proxy.key',
        'core.crt', 'core.key',
        'job_service.crt', 'job_service.key',
        'registryctl.crt', 'registryctl.key',
        'registry.crt', 'registry.key',
        'portal.crt', 'portal.key'
    }

    trivy_certs_filename = {
        'trivy_adapter.crt', 'trivy_adapter.key',
    }

    notary_certs_filename = {
        'notary_signer.crt', 'notary_signer.key',
        'notary_server.crt', 'notary_server.key'
    }

    chart_museum_filename = {
        'chartmuseum.crt',
        'chartmuseum.key'
    }

    db_certs_filename = {
        'harbor_db.crt', 'harbor_db.key'
    }

    def __init__(self, tls_enabled=False, verify_client_cert=False, tls_dir='', data_volume='', **kwargs):
        self.data_volume = data_volume
        self.verify_client_cert = verify_client_cert
        self.enabled = tls_enabled
        self.tls_dir = tls_dir
        if self.enabled:
            self.required_filenames = self.harbor_certs_filename
            if kwargs.get('with_notary'):
                self.required_filenames.update(self.notary_certs_filename)
            if kwargs.get('with_chartmuseum'):
                self.required_filenames.update(self.chart_museum_filename)
            if kwargs.get('with_trivy'):
                self.required_filenames.update(self.trivy_certs_filename)
            if not kwargs.get('external_database'):
                self.required_filenames.update(self.db_certs_filename)

    def __getattribute__(self, name: str):
        """
        Make the call like 'internal_tls.core_crt_path' possible
        """
        # only handle when enabled tls and name ends with 'path'
        if name.endswith('_path'):
            if not (self.enabled):
                return object.__getattribute__(self, name)

            name_parts = name.split('_')
            if len(name_parts) < 3:
                return object.__getattribute__(self, name)

            filename = '{}.{}'.format('_'.join(name_parts[:-2]), name_parts[-2])

            if filename in self.required_filenames:
                return os.path.join(self.data_volume, 'secret', 'tls', filename)

        return object.__getattribute__(self, name)

    def _check(self, filename: str):
        """
        Check cert and key files are correct
        """

        path = Path(os.path.join(internal_tls_dir, filename))

        if not path.exists:
            if filename == 'harbor_internal_ca.crt':
                return
            raise Exception('File {} not exist'.format(filename))

        if not path.is_file:
            raise Exception('invalid {}'.format(filename))

        # check key file permission
        if filename.endswith('.key') and not check_permission(path, mode=0o600):
            raise Exception('key file {} permission is not 600'.format(filename))

        # check certificate file
        if filename.endswith('.crt'):
            if not owner_can_read(path.stat().st_mode):
                # check owner can read cert file
                raise Exception('File {} should readable by owner'.format(filename))
            if not san_existed(path):
                # check SAN included
                if filename == 'harbor_internal_ca.crt':
                    return
                raise Exception('cert file {} should include SAN'.format(filename))

    def validate(self):
        if not self.enabled:
            # pass the validation if not enabled
            return

        if not internal_tls_dir.exists():
            raise Exception('Internal dir for tls {} not exist'.format(internal_tls_dir))

        for filename in self.required_filenames:
            self._check(filename)

    def prepare(self):
        """
        Prepare moves certs in tls file to data volume with correct permission.
        """
        if not self.enabled:
            logging.info('internal tls NOT enabled...')
            return
        original_tls_dir = get_realpath(self.tls_dir)
        if internal_tls_dir.exists():
            rmtree(internal_tls_dir)
        copytree(original_tls_dir, internal_tls_dir, symlinks=True)

        for file in internal_tls_dir.iterdir():
            if file.name.endswith('.key'):
                file.chmod(0o600)
            elif file.name.endswith('.crt'):
                file.chmod(0o644)

            if file.name in self.db_certs_filename:
                os.chown(file, PG_UID, PG_GID)
            else:
                os.chown(file, DEFAULT_UID, DEFAULT_GID)


class Metric:
    def __init__(self, enabled: bool = False, port: int = 8080, path: str = "metrics"):
        self.enabled = enabled
        self.port = port
        self.path = path

    def validate(self):
        if not port_number_valid(self.port):
            raise Exception('Port number in metrics is not valid')


class JaegerExporter:
    def __init__(self, config: dict):
        if not config:
            self.enabled = False
            return
        self.enabled = True
        self.endpoint = config.get('endpoint')
        self.username = config.get('username')
        self.password = config.get('password')
        self.agent_host = config.get('agent_host')
        self.agent_port = config.get('agent_port')

    def validate(self):
        if not self.endpoint and not self.agent_host:
            raise Exception('Jaeger Colector Endpoint or Agent host not set, must set one')
        if self.endpoint and self.agent_host:
            raise Exception('Jaeger Colector Endpoint and Agent host both set, only can set one')


class OtelExporter:
    def __init__(self, config: dict):
        if not config:
            self.enabled = False
            return
        self.enabled = True
        self.endpoint = config.get('endpoint')
        self.url_path = config.get('url_path')
        self.compression = config.get('compression') or False
        self.insecure = config.get('insecure') or False
        self.timeout = config.get('timeout') or '10s'

    def validate(self):
        if not self.endpoint:
            raise Exception('Trace endpoint not set')
        if not self.url_path:
            raise Exception('Trace url path not set')


class Trace:
    def __init__(self, config: dict):
        self.enabled = config.get('enabled') or False
        self.sample_rate = config.get('sample_rate', 1)
        self.namespace = config.get('namespace') or ''
        self.jaeger = JaegerExporter(config.get('jaeger'))
        self.otel = OtelExporter(config.get('otel'))
        self.attributes = config.get('attributes') or {}

    def validate(self):
        if not self.enabled:
            return
        if not self.jaeger.enabled and not self.otel.enabled:
            raise Exception('Trace enabled but no trace exporter set')
        elif self.jaeger.enabled and self.otel.enabled:
            raise Exception('Only can have one trace exporter at a time')
        elif self.jaeger.enabled:
            self.jaeger.validate()
        elif self.otel.enabled:
            self.otel.validate()


class PurgeUpload:
    def __init__(self, config: dict):
        if not config:
            self.enabled = False
        self.enabled = config.get('enabled')
        self.age = config.get('age') or '168h'
        self.interval = config.get('interval') or '24h'
        self.dryrun = config.get('dryrun') or False
        return

    def validate(self):
        if not self.enabled:
            return
        # age should end with h
        if not isinstance(self.age, str) or not self.age.endswith('h'):
            raise Exception('purge upload age should set with with nh, n is the number of hour')
        # interval should larger than 2h
        age = self.age[:-1]
        if not age.isnumeric() or int(age) < 2:
            raise Exception('purge upload age should set with with nh, n is the number of hour and n should not be less than 2')

        # interval should end with h
        if not isinstance(self.interval, str) or not self.interval.endswith('h'):
            raise Exception('purge upload interval should set with with nh, n is the number of hour')
        # interval should larger than 2h
        interval = self.interval[:-1]
        if not interval.isnumeric() or int(interval) < 2:
            raise Exception('purge upload interval should set with with nh, n is the number of hour and n should not beless than 2')
        return


class Cache:
    def __init__(self, config: dict):
        if not config:
            self.enabled = False
        self.enabled = config.get('enabled')
        self.expire_hours = config.get('expire_hours')

    def validate(self):
        if not self.enabled:
            return

        if not self.expire_hours.isnumeric():
            raise Exception('cache expire hours should be number')
        return
