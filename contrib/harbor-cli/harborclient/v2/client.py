from harborclient import client
from harborclient.v2 import configurations
from harborclient.v2 import jobs
from harborclient.v2 import logs
from harborclient.v2 import projects
from harborclient.v2 import repositories
from harborclient.v2 import searcher
from harborclient.v2 import statistics
from harborclient.v2 import systeminfo
from harborclient.v2 import targets
from harborclient.v2 import users


class Client(object):
    """Top-level object to access the Harbor API.

    .. warning:: All scripts and projects should not initialize this class
      directly. It should be done via `harborclient.client.Client` interface.
    """

    def __init__(self,
                 username=None,
                 password=None,
                 project=None,
                 baseurl=None,
                 insecure=False,
                 cacert=None,
                 api_version=None,
                 *argv,
                 **kwargs):
        """Initialization of Client object.

        :param str username: Username
        :param str password: Password
        :param str project: Project
        """
        self.baseurl = baseurl
        self.users = users.UserManager(self)
        self.projects = projects.ProjectManager(self)
        self.jobs = jobs.JobManager(self)
        self.repositories = repositories.RepositoryManager(self)
        self.searcher = searcher.SearchManager(self)
        self.statistics = statistics.StatisticsManager(self)
        self.logs = logs.LogManager(self)
        self.targets = targets.TargetManager(self)
        self.systeminfo = systeminfo.SystemInfoManager(self)
        self.configurations = configurations.ConfigurationManager(self)
        self.client = client._construct_http_client(
            username=username,
            password=password,
            project=project,
            baseurl=baseurl,
            insecure=insecure,
            cacert=cacert,
            api_version=api_version,
            **kwargs)

    def get_timings(self):
        return self.client.get_timings()

    def reset_timings(self):
        self.client.reset_timings()

    def authenticate(self):
        """Authenticate against the server.

        Normally this is called automatically when you first access the API,
        but you can call this method to force authentication right now.

        Returns on success; raises :exc:`exceptions.Unauthorized` if the
        credentials are wrong.
        """
        self.client.authenticate()
