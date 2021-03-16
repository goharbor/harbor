# -*- coding: utf-8 -*-

import base
import v2_swagger_client

class Retention(base.Base):
    def __init__(self):
        super(Retention,self).__init__(api_type="retention")

    def get_metadatas(self, expect_status_code = 200, **kwargs):
        metadatas, status_code, _ = self._get_client(**kwargs).get_rentenition_metadata_with_http_info()
        base._assert_status_code(expect_status_code, status_code)
        return metadatas

    def create_retention_policy(self, project_id, selector_repository="**", selector_tag="**", expect_status_code = 201, **kwargs):
        policy=v2_swagger_client.RetentionPolicy(
            algorithm='or',
            rules=[
                v2_swagger_client.RetentionRule(
                    disabled=False,
                    action="retain",
                    template="always",
                    params= {

                    },
                    scope_selectors={
                        "repository": [
                            {
                                "kind": "doublestar",
                                "decoration": "repoMatches",
                                "pattern": selector_repository
                            }
                        ]
                    },
                    tag_selectors=[
                        {
                            "kind": "doublestar",
                            "decoration": "matches",
                            "pattern": selector_tag
                        }
                    ]
                )
            ],
            trigger= {
                "kind": "Schedule",
                "settings": {
                    "cron": ""
                },
                "references": {
                }
            },
            scope= {
                "level": "project",
                "ref": project_id
            },
        )
        _, status_code, header = self._get_client(**kwargs).create_retention_with_http_info(policy)
        base._assert_status_code(expect_status_code, status_code)
        return base._get_id_from_header(header)

    def get_retention_policy(self, retention_id, expect_status_code = 200, **kwargs):
        policy, status_code, _ = self._get_client(**kwargs).get_retention_with_http_info(retention_id)
        base._assert_status_code(expect_status_code, status_code)
        return policy

    def update_retention_policy(self, retention_id, selector_repository="**", selector_tag="**", expect_status_code = 200, **kwargs):
        policy=v2_swagger_client.RetentionPolicy(
            id=retention_id,
            algorithm='or',
            rules=[
                v2_swagger_client.RetentionRule(
                    disabled=False,
                    action="retain",
                    template="always",
                    params= {

                    },
                    scope_selectors={
                        "repository": [
                            {
                                "kind": "doublestar",
                                "decoration": "repoMatches",
                                "pattern": selector_repository
                            }
                        ]
                    },
                    tag_selectors=[
                        {
                            "kind": "doublestar",
                            "decoration": "matches",
                            "pattern": selector_tag
                        }
                    ]
                )
            ],
            trigger= {
                "kind": "Schedule",
                "settings": {
                    "cron": ""
                },
                "references": {
                }
            },
            scope= {
                "level": "project",
                "ref": project_id
            },
        )
        _, status_code, _ = self._get_client(**kwargs).update_retention_with_http_info(retention_id, policy)
        base._assert_status_code(expect_status_code, status_code)

    def update_retention_add_rule(self, retention_id, selector_repository="**", selector_tag="**", with_untag="True", expect_status_code = 200, **kwargs):
        retention_rule = v2_swagger_client.RetentionRule(
                            disabled=False,
                            action="retain",
                            template="always",
                            params= {

                            },
                            scope_selectors={
                                "repository": [
                                    {
                                        "kind": "doublestar",
                                        "decoration": "repoMatches",
                                        "pattern": selector_repository
                                    }
                                ]
                            },
                            tag_selectors=[
                                {
                                    "kind": "doublestar",
                                    "decoration": "matches",
                                    "extras":'["untagged":'+with_untag+']',
                                    "pattern": selector_tag
                                }
                            ]
                        )
        client = self._get_client(**kwargs)
        policy, status_code, _ = client.get_retention_with_http_info(retention_id)
        base._assert_status_code(200, status_code)
        policy.rules.append(retention_rule)
        _, status_code, _ = client.update_retention_with_http_info(retention_id, policy)
        base._assert_status_code(expect_status_code, status_code)

    def trigger_retention_policy(self, retention_id, dry_run=False, expect_status_code = 201, **kwargs):
        _, status_code, _ = self._get_client(**kwargs).trigger_retention_execution_with_http_info(retention_id, {"dry_run":dry_run})
        base._assert_status_code(expect_status_code, status_code)

    def stop_retention_execution(self, retention_id, exec_id, expect_status_code = 200, **kwargs):
        r, status_code, _ = self._get_client(**kwargs).operate_retention_execution_with_http_info(retention_id, exec_id, {"action":"stop"})
        base._assert_status_code(expect_status_code, status_code)
        return r

    def get_retention_executions(self, retention_id, expect_status_code = 200, **kwargs):
        r, status_code, _ = self._get_client(**kwargs).list_retention_executions_with_http_info(retention_id)
        base._assert_status_code(expect_status_code, status_code)
        return r

    def get_retention_exec_tasks(self, retention_id, exec_id, expect_status_code = 200, **kwargs):
        r, status_code, _ = self._get_client(**kwargs).list_retention_tasks_with_http_info(retention_id, exec_id)
        base._assert_status_code(expect_status_code, status_code)
        return r

    def get_retention_exec_task_log(self, retention_id, exec_id, task_id, expect_status_code = 200, **kwargs):
        r, status_code, _ = self._get_client(**kwargs).get_retention_task_log_with_http_info(retention_id, exec_id, task_id)
        base._assert_status_code(expect_status_code, status_code)
        return r

    def get_retention_metadatas(self, expect_status_code = 200, **kwargs):
        r, status_code, _ = self._get_client(**kwargs).get_rentenition_metadata_with_http_info()
        base._assert_status_code(expect_status_code, status_code)
        return r
