from harborclient import base
from harborclient import exceptions as exp


class UserManager(base.Manager):
    def is_id(self, key):
        return key.isdigit()

    def get(self, id):
        """Get a user's profile."""
        return self._get("/users/%s" % id)

    def current(self):
        """Get current user info."""
        return self._get("/users/current")

    def list(self):
        """Get registered users of Harbor."""
        return self._list("/users")

    def get_id_by_name(self, name):
        users = self.list()
        for u in users:
            if u['username'] == name:
                return u['user_id']
        raise exp.NotFound("User '%s' Not Found!" % name)

    def find(self, key):
        if self.is_id(key):
            return self.get(key)
        else:
            users = self.list()
            for user in users:
                if user['username'] == key:
                    return user
        raise exp.NotFound("User '%s' Not Found!" % key)

    def create(self, username, password, email, realname=None, comment=None):
        """Creates a new user account."""
        data = {
            "username": username,
            "password": password,
            "email": email,
            "realname": realname or username,
            "comment": comment or "",
        }
        return self._create("/users", data)

    def update(self, id, realname, email, comment):
        """Update a registered user to change his profile."""
        profile = {"realname": realname,
                   "email": email,
                   "comment": comment}
        return self._update("/users/%s" % id, profile)

    def delete(self, id):
        """Mark a registered user as be removed."""
        return self._delete("/users/%s" % id)

    def change_password(self, id, old_password, new_password):
        """Change the password on a user that already exists."""
        profile = {"old_password": old_password,
                   "new_password": new_password}
        return self._update("/users/%s/password" % id, profile)

    def set_admin(self, id, is_admin):
        """Update a registered user to change to be an admin of Harbor."""
        if is_admin:
            profile = {"has_admin_role": 1}
        else:
            profile = {"has_admin_role": 0}
        return self._update("/users/%s/sysadmin" % id, profile)
