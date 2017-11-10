from harborclient import base


class LogManager(base.Manager):
    def list(self):
        """Get recent logs of the projects which the user is a member of."""
        return self._list("/logs")
