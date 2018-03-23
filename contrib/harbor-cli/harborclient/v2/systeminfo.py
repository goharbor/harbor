from harborclient import base


class SystemInfoManager(base.Manager):
    def get(self):
        """Get general system info."""
        return self._get("/systeminfo")

    def get_volumes(self):
        """Get system volume info (total/free size)."""
        return self._get("/systeminfo/volumes")

    def get_cert(self):
        """Get default root certificate under OVA deployment."""
        return self._get("/systeminfo/getcert")
