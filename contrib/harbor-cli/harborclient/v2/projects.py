from harborclient import base
from harborclient import exceptions as exp


class ProjectManager(base.Manager):
    def is_id(self, key):
        return key.isdigit()

    def get(self, id):
        """Return specific project detail infomation."""
        return self._get("/projects/%s" % id)

    def list(self):
        """List projects."""
        return self._list("/projects")

    def get_id_by_name(self, name):
        """Return specific project detail infomation by name."""
        projects = self.list()
        for p in projects:
            if p['name'] == name:
                return p['project_id']
        raise exp.NotFound("Project '%s' not Found." % name)

    def get_name_by_id(self, id):
        """Return specific project detail infomation by id."""
        projects = self.list()
        for p in projects:
            if p['project_id'] == id:
                return p['name']
        raise exp.NotFound("Project '%s' not Found." % id)

    def create(self, name, public=True):
        """Create a new project."""
        project = {"project_name": name, "public": 1 if public else 0}
        return self._create("/projects", project)

    def delete(self, id):
        """Delete project by id."""
        return self._delete("/projects/%s" % id)

    def get_members(self, id):
        """Return a project's relevant role members."""
        return self._list("/projects/%s/members/" % id)
