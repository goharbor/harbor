import base
from client.rest import ApiException

class Chart(base.Base, object):
    def __init__(self):
        super(Chart,self).__init__(api_type = "chart")

    def upload_chart(self, repository, chart, prov = None, expect_status_code = 201, **kwargs):
        client = self._get_client(**kwargs)
        try:
            _, status_code, _ = client.chartrepo_repo_charts_post_with_http_info(repository, chart)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
        else:
            base._assert_status_code(expect_status_code, status_code)
            base._assert_status_code(201, status_code)

    def get_charts(self, repository, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        try:
            body, status_code, _ = client.chartrepo_repo_charts_get_with_http_info(repository)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            return []
        else:
            base._assert_status_code(expect_status_code, status_code)
            base._assert_status_code(200, status_code)
            return body

    def chart_should_exist(self, repository, chart_name, expect_status_code = 200, **kwargs):
        charts_data = self.get_charts(repository, expect_status_code = expect_status_code, **kwargs)
        for chart in charts_data:
            if chart.name == chart_name:
                return True
        if expect_status_code == 200:
            raise Exception(r"Chart {} does not exist in project {}.".format(chart_name, repository))

    def delete_chart_with_version(self, repository, chart_name, version, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        try:
            _, status_code, _ = client.chartrepo_repo_charts_name_version_delete_with_http_info(repository, chart_name, version)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
        else:
            base._assert_status_code(expect_status_code, status_code)
            base._assert_status_code(200, status_code)