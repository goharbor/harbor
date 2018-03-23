import pbr.version

from harborclient import api_versions

__version__ = pbr.version.VersionInfo('python-harborclient').version_string()

API_MIN_VERSION = api_versions.APIVersion("2.0")
# The max version should be the latest version that is supported in the client,
# not necessarily the latest that the server can provide. This is only bumped
# when client supported the max version, and bumped sequentially, otherwise
# the client may break due to server side new version may include some
# backward incompatible change.
API_MAX_VERSION = api_versions.APIVersion("2.0")
