import logging
import os
import pkgutil
import re

import harborclient
from harborclient import exceptions

LOG = logging.getLogger(__name__)
_type_error_msg = "'%(other)s' should be an instance of '%(cls)s'"


if not LOG.handlers:
    LOG.addHandler(logging.StreamHandler())


class APIVersion(object):
    """This class represents an API Version Request.

    This class provides convenience methods for manipulation
    and comparison of version numbers that we need to do to
    implement microversions.
    """

    def __init__(self, version_str=None):
        """Create an API version object.

        :param version_str: String representation of APIVersionRequest.
                            Correct format is 'X.Y', where 'X' and 'Y'
                            are int values. None value should be used
                            to create Null APIVersionRequest, which is
                            equal to 0.0
        """
        self.ver_major = 0
        self.ver_minor = 0

        if version_str is not None:
            match = re.match(r"^([1-9]\d*)\.([1-9]\d*|0|latest)$", version_str)
            if match:
                self.ver_major = int(match.group(1))
                if match.group(2) == "latest":
                    # NOTE(andreykurilin): Infinity allows to easily determine
                    # latest version and doesn't require any additional checks
                    # in comparison methods.
                    self.ver_minor = float("inf")
                else:
                    self.ver_minor = int(match.group(2))
            else:
                msg = ("Invalid format of client version '%s'. "
                       "Expected format 'X.Y', where X is a major part and Y "
                       "is a minor part of version.") % version_str
                raise exceptions.UnsupportedVersion(msg)

    def __str__(self):
        """Debug/Logging representation of object."""
        if self.is_latest():
            return "Latest API Version Major: %s" % self.ver_major
        return ("API Version Major: %s, Minor: %s" % (self.ver_major,
                                                      self.ver_minor))

    def __repr__(self):
        if self.is_null():
            return "<APIVersion: null>"
        else:
            return "<APIVersion: %s>" % self.get_string()

    def is_null(self):
        return self.ver_major == 0 and self.ver_minor == 0

    def is_latest(self):
        return self.ver_minor == float("inf")

    def __lt__(self, other):
        if not isinstance(other, APIVersion):
            raise TypeError(
                _type_error_msg % {"other": other,
                                   "cls": self.__class__})

        return ((self.ver_major, self.ver_minor) <
                (other.ver_major, other.ver_minor))

    def __eq__(self, other):
        if not isinstance(other, APIVersion):
            raise TypeError(
                _type_error_msg % {"other": other,
                                   "cls": self.__class__})

        return ((self.ver_major, self.ver_minor) == (other.ver_major,
                                                     other.ver_minor))

    def __gt__(self, other):
        if not isinstance(other, APIVersion):
            raise TypeError(
                _type_error_msg % {"other": other,
                                   "cls": self.__class__})

        return ((self.ver_major, self.ver_minor) >
                (other.ver_major, other.ver_minor))

    def __le__(self, other):
        return self < other or self == other

    def __ne__(self, other):
        return not self.__eq__(other)

    def __ge__(self, other):
        return self > other or self == other

    def matches(self, min_version, max_version):
        """Matches the version object.

        Returns whether the version object represents a version
        greater than or equal to the minimum version and less than
        or equal to the maximum version.

        :param min_version: Minimum acceptable version.
        :param max_version: Maximum acceptable version.
        :returns: boolean

        If min_version is null then there is no minimum limit.
        If max_version is null then there is no maximum limit.
        If self is null then raise ValueError
        """

        if self.is_null():
            raise ValueError("Null APIVersion doesn't support 'matches'.")
        if max_version.is_null() and min_version.is_null():
            return True
        elif max_version.is_null():
            return min_version <= self
        elif min_version.is_null():
            return self <= max_version
        else:
            return min_version <= self <= max_version

    def get_string(self):
        """Version string representation.

        Converts object to string representation which if used to create
        an APIVersion object results in the same version.
        """
        if self.is_null():
            raise ValueError("Null APIVersion cannot be converted to string.")
        elif self.is_latest():
            return "%s.%s" % (self.ver_major, "latest")
        return "%s.%s" % (self.ver_major, self.ver_minor)


class VersionedMethod(object):
    def __init__(self, name, start_version, end_version, func):
        """Versioning information for a single method

        :param name: Name of the method
        :param start_version: Minimum acceptable version
        :param end_version: Maximum acceptable_version
        :param func: Method to call

        Minimum and maximums are inclusive
        """
        self.name = name
        self.start_version = start_version
        self.end_version = end_version
        self.func = func

    def __str__(self):
        return ("Version Method %s: min: %s, max: %s" %
                (self.name, self.start_version, self.end_version))

    def __repr__(self):
        return "<VersionedMethod %s>" % self.name


def get_available_major_versions():
    # NOTE(andreykurilin): available clients version should not be
    # hardcoded, so let's discover them.
    matcher = re.compile(r"v[0-9]*$")
    submodules = pkgutil.iter_modules([os.path.dirname(__file__)])
    available_versions = [
        name[1:] for loader, name, ispkg in submodules if matcher.search(name)
    ]

    return available_versions


def check_major_version(api_version):
    """Checks major part of ``APIVersion`` obj is supported.

    :raises harborclient.exceptions.UnsupportedVersion: if major part is not
                                                      supported
    """
    available_versions = get_available_major_versions()
    if (not api_version.is_null() and
            str(api_version.ver_major) not in available_versions):
        if len(available_versions) == 1:
            msg = ("Invalid client version '%(version)s'. "
                   "Major part should be '%(major)s'") % {
                       "version": api_version.get_string(),
                       "major": available_versions[0]}
        else:
            msg = ("Invalid client version '%(version)s'. "
                   "Major part must be one of: '%(major)s'") % {
                       "version": api_version.get_string(),
                       "major": ", ".join(available_versions)}
        raise exceptions.UnsupportedVersion(msg)


def get_api_version(version_string):
    """Returns checked APIVersion object"""
    version_string = str(version_string)
    api_version = APIVersion(version_string)
    check_major_version(api_version)
    return api_version


def _get_server_version_range(client):
    version = client.versions.get_current()

    if not hasattr(version, 'version') or not version.version:
        return APIVersion(), APIVersion()

    return APIVersion(version.min_version), APIVersion(version.version)


def discover_version(client, requested_version):
    """Discover most recent version supported by API and client.

    Checks ``requested_version`` and returns the most recent version
    supported by both the API and the client.

    :param client: client object
    :param requested_version: requested version represented by APIVersion obj
    :returns: APIVersion
    """
    server_start_version, server_end_version = _get_server_version_range(
        client)

    if (not requested_version.is_latest() and
            requested_version != APIVersion('2.0')):
        if server_start_version.is_null() and server_end_version.is_null():
            raise exceptions.UnsupportedVersion(
                ("Server doesn't support microversions"))
        if not requested_version.matches(server_start_version,
                                         server_end_version):
            raise exceptions.UnsupportedVersion(
                ("The specified version isn't supported by server. The valid "
                 "version range is '%(min)s' to '%(max)s'") % {
                     "min": server_start_version.get_string(),
                     "max": server_end_version.get_string()})
        return requested_version

    if server_start_version.is_null() and server_end_version.is_null():
        return APIVersion('2.0')
    elif harborclient.API_MIN_VERSION > server_end_version:
        raise exceptions.UnsupportedVersion(
            ("Server version is too old. The client valid version range is "
             "'%(client_min)s' to '%(client_max)s'. The server valid version "
             "range is '%(server_min)s' to '%(server_max)s'.") % {
                 'client_min': harborclient.API_MIN_VERSION.get_string(),
                 'client_max': harborclient.API_MAX_VERSION.get_string(),
                 'server_min': server_start_version.get_string(),
                 'server_max': server_end_version.get_string()})
    elif harborclient.API_MAX_VERSION < server_start_version:
        raise exceptions.UnsupportedVersion(
            ("Server version is too new. The client valid version range is "
             "'%(client_min)s' to '%(client_max)s'. The server valid version "
             "range is '%(server_min)s' to '%(server_max)s'.") % {
                 'client_min': harborclient.API_MIN_VERSION.get_string(),
                 'client_max': harborclient.API_MAX_VERSION.get_string(),
                 'server_min': server_start_version.get_string(),
                 'server_max': server_end_version.get_string()})
    elif harborclient.API_MAX_VERSION <= server_end_version:
        return harborclient.API_MAX_VERSION
    elif server_end_version < harborclient.API_MAX_VERSION:
        return server_end_version
