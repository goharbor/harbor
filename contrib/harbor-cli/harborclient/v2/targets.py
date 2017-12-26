from harborclient import base


class TargetManager(base.Manager):

    def list(self, name=None):
        """List filters targets by name."""
        if name:
            return self._list("/targets?name=%s" % name)
        return self._list("/targets")

    def ping(self, id):
        """Ping validates target."""
        return self._create("/targets/%s/ping" % id)

    def list_policies(self, id):
        """List the target relevant policies."""
        return self._list("/targets/%s/policies" % id)
