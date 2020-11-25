# -*- coding: utf-8 -*-

import base
import swagger_client
from swagger_client.rest import ApiException

class Tag_Immutability(base.Base):
    def create_tag_immutability_policy_rule(self, project_id, selector_repository_decoration = "repoMatches",
                                            selector_repository="**", selector_tag_decoration = "matches",
                                            selector_tag="**", expect_status_code = 201, **kwargs):
        #repoExcludes,excludes
        client = self._get_client(**kwargs)
        retention_rule = swagger_client.RetentionRule(
                    action="immutable",
                    template="immutable_template",
                    priority = 0,
                    scope_selectors={
                        "repository": [
                            {
                                "kind": "doublestar",
                                "decoration": selector_repository_decoration,
                                "pattern": selector_repository
                            }
                        ]
                    },
                    tag_selectors=[
                        {
                            "kind": "doublestar",
                            "decoration": selector_tag_decoration,
                            "pattern": selector_tag
                        }
                    ]
                )
        try:
            _, status_code, header = client.projects_project_id_immutabletagrules_post_with_http_info(project_id, retention_rule)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
        else:
            base._assert_status_code(expect_status_code, status_code)
            base._assert_status_code(201, status_code)
            return base._get_id_from_header(header)

    def list_tag_immutability_policy_rules(self, project_id, **kwargs):
        client = self._get_client(**kwargs)
        return client.projects_project_id_immutabletagrules_get(project_id)

    def get_rule(self, project_id, rule_id, **kwargs):
        rules = self.list_tag_immutability_policy_rules(project_id, **kwargs)
        for r in rules:
            if r.id == rule_id:
                return r
        return None

    def update_tag_immutability_policy_rule(self, project_id, rule_id, selector_repository_decoration = None,
                                            selector_repository=None, selector_tag_decoration = None,
                                            selector_tag=None, disabled = None, expect_status_code = 200, **kwargs):
        rule = self.get_rule( project_id, rule_id,**kwargs)
        if selector_repository_decoration:
            rule.scope_selectors["repository"][0].decoration = selector_repository_decoration
        if selector_repository:
            rule.scope_selectors["repository"][0].pattern = selector_repository
        if selector_tag_decoration:
            rule.tag_selectors[0].decoration = selector_tag_decoration
        if selector_tag:
            rule.tag_selectors[0].pattern = selector_tag
        if disabled is not None:
            rule.disabled = disabled
        client = self._get_client(**kwargs)
        try:
            _, status_code, header = client.projects_project_id_immutabletagrules_id_put_with_http_info(project_id, rule_id, rule)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
        else:
            base._assert_status_code(expect_status_code, status_code)
            base._assert_status_code(200, status_code)
            return base._get_id_from_header(header)

    def create_rule(self, project_id, selector_repository_decoration = "repoMatches", selector_repository="**",
                                      selector_tag_decoration = "matches", selector_tag="**",
                                      expect_status_code = 201, disabled = False, **kwargs):
        rule_id = self.create_tag_immutability_policy_rule(project_id, selector_repository_decoration = selector_repository_decoration,
                                                            selector_repository = selector_repository,
                                                            selector_tag_decoration = selector_tag_decoration,
                                                            selector_tag = selector_tag, expect_status_code = expect_status_code, **kwargs)
        if expect_status_code != 201:
            return
        self.update_tag_immutability_policy_rule(project_id, rule_id, selector_repository_decoration = selector_repository_decoration,
                                                 selector_repository = selector_repository, selector_tag_decoration = selector_tag_decoration,
                                                 selector_tag = selector_tag,  disabled = disabled, expect_status_code = 200, **kwargs)
        return rule_id

