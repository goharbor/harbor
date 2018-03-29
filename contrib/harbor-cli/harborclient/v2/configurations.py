from harborclient import base


class ConfigurationManager(base.Manager):
    def get(self):
        """Get system configurations."""
        return self._get("/configurations")
