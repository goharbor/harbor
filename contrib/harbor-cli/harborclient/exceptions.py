"""
Exception definitions.
"""


class UnsupportedVersion(Exception):
    """Unsupport API version.

    Indicates that the user is trying to use an unsupported version of the API.
    """
    pass


class UnsupportedAttribute(AttributeError):
    """Unsupport attribute

    Indicates that the user is trying to transmit the argument to a method,
    which is not supported by selected version.
    """

    def __init__(self, argument_name, start_version, end_version=None):
        if end_version:
            self.message = (
                "'%(name)s' argument is only allowed for microversions "
                "%(start)s - %(end)s." % {
                    "name": argument_name,
                    "start": start_version,
                    "end": end_version
                })
        else:
            self.message = (
                "'%(name)s' argument is only allowed since microversion "
                "%(start)s." % {
                    "name": argument_name,
                    "start": start_version
                })


class CommandError(Exception):
    pass


class AuthorizationFailure(Exception):
    pass


class ClientException(Exception):
    """The base exception class for all exceptions this library raises."""
    message = 'Unknown Error'

    def __init__(self,
                 code,
                 message=None,
                 details=None,
                 request_id=None,
                 url=None,
                 method=None):
        self.code = code
        self.message = message or self.__class__.message
        self.details = details
        self.request_id = request_id
        self.url = url
        self.method = method

    def __str__(self):
        formatted_string = "%s (HTTP %s)" % (self.message, self.code)
        if self.request_id:
            formatted_string += " (Request-ID: %s)" % self.request_id

        return formatted_string


class RetryAfterException(ClientException):
    """Retry exception

    The base exception class for ClientExceptions that use Retry-After header.
    """

    def __init__(self, *args, **kwargs):
        try:
            self.retry_after = int(kwargs.pop('retry_after'))
        except (KeyError, ValueError):
            self.retry_after = 0

        super(RetryAfterException, self).__init__(*args, **kwargs)


class BadRequest(ClientException):
    """HTTP 400 - Bad request: you sent some malformed data."""
    http_status = 400
    message = "Bad request"


class Unauthorized(ClientException):
    """HTTP 401 - Unauthorized: bad credentials."""
    http_status = 401
    message = "Unauthorized"


class Forbidden(ClientException):
    """HTTP 403 - Forbidden

    HTTP 403 - Forbidden: your credentials don't give you access to this
    resource.
    """
    http_status = 403
    message = "Forbidden"


class NotFound(ClientException):
    """HTTP 404 - Not found"""
    http_status = 404
    message = "Not found"


class MethodNotAllowed(ClientException):
    """HTTP 405 - Method Not Allowed"""
    http_status = 405
    message = "Method Not Allowed"


class NotAcceptable(ClientException):
    """HTTP 406 - Not Acceptable"""
    http_status = 406
    message = "Not Acceptable"


class Conflict(ClientException):
    """HTTP 409 - Conflict"""
    http_status = 409
    message = "Conflict"


class OverLimit(RetryAfterException):
    """HTTP 413 - Over limit

    You're over the API limits for this time period.
    """
    http_status = 413
    message = "Over limit"


class RateLimit(RetryAfterException):
    """HTTP 429 - Rate limit

    you've sent too many requests for this time period.
    """
    http_status = 429
    message = "Rate limit"


# NotImplemented is a python keyword.
class HTTPNotImplemented(ClientException):
    """HTTP 501 - Not Implemented

    the server does not support this operation.
    """
    http_status = 501
    message = "Not Implemented"


# In Python 2.4 Exception is old-style and thus doesn't have a __subclasses__()
# so we can do this:
#     _code_map = dict((c.http_status, c)
#                      for c in ClientException.__subclasses__())
#
# Instead, we have to hardcode it:
_error_classes = [
    BadRequest, Unauthorized, Forbidden, NotFound, MethodNotAllowed,
    NotAcceptable, Conflict, OverLimit, RateLimit, HTTPNotImplemented
]
_code_map = dict((c.http_status, c) for c in _error_classes)


def from_response(response, body, url, method=None):
    """Extract exception from response

    Return an instance of an ClientException or subclass baseda
    on a requests response.

    Usage::

        resp, body = requests.request(...)
        if resp.status_code != 200:
            raise exception_from_response(resp, rest.text)
    """
    cls = _code_map.get(response.status_code, ClientException)
    kwargs = {
        'code': response.status_code,
        'method': method,
        'url': url,
        'message': body.strip(),
    }
    return cls(**kwargs)
