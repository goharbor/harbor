from harborclient import base


class StatisticsManager(base.Manager):
    def list(self):
        """Get projects number and repositories number relevant to the user."""
        return self._list("/statistics")
