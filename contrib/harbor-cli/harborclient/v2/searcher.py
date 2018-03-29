from harborclient import base


class SearchManager(base.Manager):
    def search(self, query):
        """Search for projects and repositories."""
        return self.api.client.get("/search?q=%s" % query)
