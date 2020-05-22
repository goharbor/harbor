import base

class Chart(base.Base, object):
    def __init__(self):
        super(Chart,self).__init__(api_type = "chart")

    def upload_chart(self, repository, chart, prov = None, expect_status_code = 201, **kwargs):
        client = self._get_client(**kwargs)
        _, status_code, _ = client.chartrepo_repo_charts_post_with_http_info(repository, chart)
        base._assert_status_code(expect_status_code, status_code)

    def get_charts(self, repository, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        body, status_code, _ = client.chartrepo_repo_charts_get_with_http_info(repository)
        base._assert_status_code(expect_status_code, status_code)
        return body

    def chart_should_exist(self, repository, chart_name, **kwargs):
        charts_data = self.get_charts(repository, **kwargs)
        print "charts_data:", charts_data
        for chart in charts_data:
            if chart.name == chart_name:
                return True
        raise Exception(r"Chart {} does not exist in project {}.".format(chart_name, repository))

    def delete_chart_with_version(self, repository, chart_name, version, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        _, status_code, _ = client.chartrepo_repo_charts_name_version_delete_with_http_info(repository, chart_name, version)
        base._assert_status_code(expect_status_code, status_code)