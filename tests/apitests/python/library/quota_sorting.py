import base
from v2_swagger_client.rest import ApiException

class QuotaSorting(base.Base):
    def __init__(self):
        super(QuotaSorting,self).__init__(api_type="quota")

    def list_quotas_with_sorting(self, expect_status_code=200, **kwargs):
        params = {}
        if "sort" in kwargs:
            params["sort"] = kwargs["sort"]
        if "reference" in kwargs:
            params["reference"] = kwargs["reference"]

        try:
            resp_data, status_code, _ = self._get_client(**kwargs).list_quotas_with_http_info(**params)
        except ApiException as e:
            raise Exception(r"Error out with exception. Exception status: {}; exception reason: {}; exception body: {}", e.status, e.reason, e.body)
        base._assert_status_code(expect_status_code, status_code)

        return resp_data
