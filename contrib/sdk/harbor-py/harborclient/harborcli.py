#!/usr/bin/python

import os, sys, argparse
import numpy as np
import harborsdk


class HarborCli(object):
    def __init__(self):
        if ('HARBOR_HOSTNAME' in os.environ) and ('HARBOR_USER' in os.environ) and ('HARBOR_PASSWORD' in os.environ):
            host = os.environ['HARBOR_HOSTNAME']
            user = os.environ['HARBOR_USER']
            pwd = os.environ['HARBOR_PASSWORD']
            protocol = 'http'
            if 'HARBOR_URL_PROTOCOL' in os.environ:
                protocol = os.environ['HARBOR_URL_PROTOCOL']
            self.sdk = harborsdk.HarborSdk(host, user, pwd, protocol)
            self.sdk.login()
        else:
            print "Harbor environments haven't been set, please export your Harbor environments first."
            sys.exit()

    def __del__(self):
        self.sdk.logout()


def main():
    argv = sys.argv[1:]
    desc = """
    project:
        check-project           Check if project exist
        create-project          Create a new project
        create-project-member   Create a new project member
        delete-project          Delete a project
        delete-project-member   Delete a project member
        get-project             Get a project
        get-project-logs        Get project logs
        get-project-members     Get project members
        get-project-role        Get a user's project role
        get-projects            Get projects or project
        toggle-project          Toggle a project to public or not
        update-project-role     Update a user's project role
        
    user:
        create-user             Create a new user
        delete-user             delete a user
        get-current-user        Get current user info
        get-user                Get user by id
        get-users               Get users, only be used by administrator
        toggle-admin            Toggle a user to admin or not
        update-password         Update user's password
        update-user             Update user profile, only email, realname, comment can be changed
        
    repository:
        create-scan-job         Create a job to call Clair API to scan image, must be admin
        delete-repo             Delete a repo
        delete-repo-tag         Delete a tag of repo
        get-repo-manifest       Get manifest of a specified repo and tag
        get-repos               Get repos with relevant projectID or repo name
        get-repo-signatures     Get signatures of a repo, Harbor should be installed with notary
        get-repo-tag            Get a tag of the repo
        get-repo-tags           Get all tags of a repo 
        get-scan-detail         Get detail info of the scan job
        get-top-repos           Get most popular public repos   
    
    job:
        delete-job              Delete a job
        get-jobs                Get jobs according to policy and repo
        get-job-logs            Get a job's logs
        
    policy:
        create-policy           Create a new policy
        enable-policy           Enable a policy or disable
        get-policy              Get a replication policy
        get-policies            Get replication policies    
        update-policy           Update a policy
    
    target:
        create-target           Create a new replication target
        delete-target           Delete a replication target
        get-target              Get a replication target
        get-target-policies     Get a target's policies
        get-targets             Get replication targets
        ping-target             Ping target with id
        test-target             Test connection with target url
        update-target           Update a replication target
    
    system:
        get-cert                Get default root cert under OVA deployment
        get-system-info         Get system info
        get-system-volumes      Get system volumes
        
    ldap:
        import-ldap-users       Import ldap users according to system configurations
        ping-ldap               Ping ldap service, if ldap url not provided, will load from system configurations 
        search-ldap-users       Search ldap users
        
    configuration:
        get-config              Get configurations of Harbor
        reset-config            Reset configurations of Harbor
        update-config           Update configurations of Harbor
    others:
        logs                    Get recent logs of the projects which the user is a member of
        ping-email              Ping email server, if settings not provided, will load from system configurations
        search                  Search for project or repo
        statistics              Get statistic data relevant to the user
        sync-registry           Sync repos from registry to DB
        
    """
    parser = argparse.ArgumentParser(prog='harborcli', description='Harbor Command Line Tool',
                                     formatter_class=argparse.RawDescriptionHelpFormatter)
    subparsers = parser.add_subparsers(title='subcommands', description=desc)

    # search
    parser_search = subparsers.add_parser('search')
    parser_search.add_argument('q', help='project or repo name')
    parser_search.set_defaults(func=search)

    # get projects
    parser_get_projects = subparsers.add_parser('get-projects')
    parser_get_projects.add_argument('-n', '--name', help='project name')
    parser_get_projects.add_argument('-i', '--public', type=int, choices=[0, 1], default=0,
                                     help='0-private, 1-public, default 0')
    parser_get_projects.add_argument('-o', '--owner', help='name of the project owner')
    parser_get_projects.add_argument('-p', '--page', type=int, default=1, help='page number, default 1')
    parser_get_projects.add_argument('-s', '--page_size', type=int, default=10, help='page size, default 10')
    parser_get_projects.set_defaults(func=get_projects)

    # check project 
    parser_check_project = subparsers.add_parser('check-project')
    parser_check_project.add_argument('project_name', help='project name')
    parser_check_project.set_defaults(func=check_project_exist)

    # create project 
    parser_create_project = subparsers.add_parser('create-project')
    parser_create_project.add_argument('project_name', help='project name')
    parser_create_project.add_argument('public', type=int, choices=[0, 1], help='0-private, 1-public')
    parser_create_project.add_argument('-e', '--enable_content_trust', type=bool, help='enable content trust or not')
    parser_create_project.add_argument('-p', '--prevent_vulnerable_images_from_running', type=bool,
                                       help='prevent vulnerable images or not')
    parser_create_project.add_argument('-s', '--prevent_vulnerable_images_from_running_severity',
                                       help='prevent severity')
    parser_create_project.add_argument('-a', '--automatically_scan_images_on_push', type=bool, help='auto scan or not')
    parser_create_project.set_defaults(func=create_project)

    # delete project 
    parser_delete_project = subparsers.add_parser('delete-project')
    parser_delete_project.add_argument('project_id', type=long, help='project id')
    parser_delete_project.set_defaults(func=delete_project)

    # get project
    parser_get_project = subparsers.add_parser('get-project')
    parser_get_project.add_argument('project_id', type=long, help='project id')
    parser_get_project.set_defaults(func=get_project)

    # toggle project publicity
    parser_toggle_project = subparsers.add_parser('toggle-project')
    parser_toggle_project.add_argument('project_id', type=long, help='project id')
    parser_toggle_project.add_argument('public', type=int, choices=[0, 1], help='0-private, 1-public')
    parser_toggle_project.set_defaults(func=toggle_project)

    # get project logs
    parser_get_project_logs = subparsers.add_parser('get-project-logs')
    parser_get_project_logs.add_argument('project_id', type=long, help='project id')
    parser_get_project_logs.add_argument('-u', '--username', help='user name')
    parser_get_project_logs.add_argument('-n', '--repository', help='repo name')
    parser_get_project_logs.add_argument('-t', '--tag', help='repo tag')
    parser_get_project_logs.add_argument('-o', '--operation', help='user operation, e.g. push, pull... ')
    parser_get_project_logs.add_argument('-b', '--begin_timestamp', type=np.int64, help='begin time of logs')
    parser_get_project_logs.add_argument('-e', '--end_timestamp', type=np.int64, help='end time of logs')
    parser_get_project_logs.add_argument('-p', '--page', type=int, default=1, help='page number, default 1')
    parser_get_project_logs.add_argument('-s', '--page_size', type=int, default=10, help='page size, default 10')
    parser_get_project_logs.set_defaults(func=get_project_logs)

    # get project members
    parser_get_project_members = subparsers.add_parser('get-project-members')
    parser_get_project_members.add_argument('project_id', type=long, help='project id')
    parser_get_project_members.set_defaults(func=get_project_members)

    # create project member
    parser_create_project_member = subparsers.add_parser('create-project-member')
    parser_create_project_member.add_argument('project_id', type=long, help='project id')
    parser_create_project_member.add_argument('role', type=int, help='project role')
    parser_create_project_member.add_argument('username', help='user name')
    parser_create_project_member.set_defaults(func=create_project_member)

    # get project role
    parser_get_project_role = subparsers.add_parser('get-project-role')
    parser_get_project_role.add_argument('project_id', type=long, help='project id')
    parser_get_project_role.add_argument('user_id', type=int, help='user id')
    parser_get_project_role.set_defaults(func=get_project_role)

    # update project role
    parser_update_project_role = subparsers.add_parser('update-project-role')
    parser_update_project_role.add_argument('project_id', type=long, help='project id')
    parser_update_project_role.add_argument('user_id', type=int, help='user id')
    parser_update_project_role.add_argument('role', type=int, help='project role')
    parser_update_project_role.add_argument('username', help='user name')
    parser_update_project_role.set_defaults(func=update_project_role)

    # delete project member
    parser_delete_project_member = subparsers.add_parser('delete-project-member')
    parser_delete_project_member.add_argument('project_id', type=long, help='project id')
    parser_delete_project_member.add_argument('user_id', type=int, help='user id')
    parser_delete_project_member.set_defaults(func=delete_project_member)

    # statistics
    parser_statistics = subparsers.add_parser('statistics')
    parser_statistics.set_defaults(func=statistics)

    # get users
    parser_get_users = subparsers.add_parser('get-users')
    parser_get_users.add_argument('-u', '--username', help='user name')
    parser_get_users.add_argument('-e', '--email', help='email address')
    parser_get_users.add_argument('-p', '--page', type=int, default=1, help='page number, default 1')
    parser_get_users.add_argument('-s', '--page_size', type=int, default=10, help='page size, default 10')
    parser_get_users.set_defaults(func=get_users)

    # create user
    parser_create_user = subparsers.add_parser('create-user')
    parser_create_user.add_argument('username', help='user name, can not contain special characters, max length 20')
    parser_create_user.add_argument('email', help='valid email address like name@example.com')
    parser_create_user.add_argument('password',
                                    help='at least 8 characters, at least 1 uppercase, 1 lowercase, 1 number')
    parser_create_user.add_argument('realname', help='max length 20')
    parser_create_user.add_argument('-c', '--comment', help='comment info')
    parser_create_user.set_defaults(func=create_user)

    # get current user
    parser_get_current_user = subparsers.add_parser('get-current-user')
    parser_get_current_user.set_defaults(func=get_current_user)

    # get user by id
    parser_get_user = subparsers.add_parser('get-user')
    parser_get_user.add_argument('user_id', type=int, help='user id')
    parser_get_user.set_defaults(func=get_user)

    # update user
    parser_update_user = subparsers.add_parser('update-user')
    parser_update_user.add_argument('user_id', type=int, help='user id')
    parser_update_user.add_argument('-e', '--email', help='valid email address')
    parser_update_user.add_argument('-n', '--realname', help='real name')
    parser_update_user.add_argument('-c', '--comment', help='comment message')
    parser_update_user.set_defaults(func=update_user)

    # delete a user
    parser_delete_user = subparsers.add_parser('delete-user')
    parser_delete_user.add_argument('user_id', type=int, help='user id')
    parser_delete_user.set_defaults(func=delete_user)

    # update password
    parser_update_password = subparsers.add_parser('update-password')
    parser_update_password.add_argument('user_id', type=int, help='user id')
    parser_update_password.add_argument('old_password', help='old password')
    parser_update_password.add_argument('new_password', help='new password')
    parser_update_password.set_defaults(func=update_password)

    # toggle admin
    parser_toggle_admin = subparsers.add_parser('toggle-admin')
    parser_toggle_admin.add_argument('user_id', type=int, help='user id')
    parser_toggle_admin.add_argument('has_admin_role', type=int, choices=[0, 1], help='0-common user, 1-admin')
    parser_toggle_admin.set_defaults(func=toggle_admin)

    # get repos
    parser_get_repos = subparsers.add_parser('get-repos')
    parser_get_repos.add_argument('project_id', type=int, help='project id')
    parser_get_repos.add_argument('-q', '--repo_name', help='repo name')
    parser_get_repos.add_argument('-p', '--page', type=int, default=1, help='page number, default 1')
    parser_get_repos.add_argument('-s', '--page_size', type=int, default=10, help='page size, default 10')
    parser_get_repos.set_defaults(func=get_repos)

    # delete repo
    parser_delete_repo = subparsers.add_parser('delete-repo')
    parser_delete_repo.add_argument('repo_name', help='e.g. library/hello-world')
    parser_delete_repo.set_defaults(func=delete_repo)

    # get repo tag
    parser_get_repo_tag = subparsers.add_parser('get-repo-tag')
    parser_get_repo_tag.add_argument('repo_name', help='repo name')
    parser_get_repo_tag.add_argument('tag', help='tag of the repo')
    parser_get_repo_tag.set_defaults(func=get_repo_tag)

    # delete repo tag
    parser_delete_repo_tag = subparsers.add_parser('delete-repo-tag')
    parser_delete_repo_tag.add_argument('repo_name', help='e.g. library/hello-world')
    parser_delete_repo_tag.add_argument('tag', help='tag of a repo')
    parser_delete_repo_tag.set_defaults(func=delete_repo_tag)

    # get repo tags
    parser_get_repo_tags = subparsers.add_parser('get-repo-tags')
    parser_get_repo_tags.add_argument('repo_name', help='e.g. library/hello-world')
    parser_get_repo_tags.set_defaults(func=get_repo_tags)

    # get repo manifest
    parser_get_repo_manifest = subparsers.add_parser('get-repo-manifest')
    parser_get_repo_manifest.add_argument('repo_name', help='e.g. library/hello-world')
    parser_get_repo_manifest.add_argument('tag', help='tag of a repo')
    parser_get_repo_manifest.add_argument('-v', '--version', choices=['v1', 'v2'], default='v2',
                                          help='manifest version')
    parser_get_repo_manifest.set_defaults(func=get_repo_manifest)

    # create scan job
    parser_create_scan_job = subparsers.add_parser('create-scan-job')
    parser_create_scan_job.add_argument('repo_name', help='e.g. library/hello-world')
    parser_create_scan_job.add_argument('tag', help='repo tag')
    parser_create_scan_job.set_defaults(func=create_scan_job)

    # get scan detail
    parser_get_scan_detail = subparsers.add_parser('get-scan-detail')
    parser_get_scan_detail.add_argument('repo_name', help='e.g. library/hello-world')
    parser_get_scan_detail.add_argument('tag', help='repo tag')
    parser_get_scan_detail.set_defaults(func=get_scan_detail)

    # get repo signatures
    parser_get_repo_sigs = subparsers.add_parser('get-repo-signatures')
    parser_get_repo_sigs.add_argument('repo_name', help='e.g. library/hello-world')
    parser_get_repo_sigs.set_defaults(func=get_repo_signatures)

    # get top repos
    parser_get_top_repos = subparsers.add_parser('get-top-repos')
    parser_get_top_repos.add_argument('-c', '--count', type=int, default=10, help='the number of top repos, default 10')
    parser_get_top_repos.set_defaults(func=get_top_repos)

    # get logs
    parser_logs = subparsers.add_parser('logs')
    parser_logs.add_argument('-u', '--username', help='user name of the operator')
    parser_logs.add_argument('-n', '--repository', help='name of the repository')
    parser_logs.add_argument('-t', '--tag', help='name of tag')
    parser_logs.add_argument('-o', '--operation', help='the operation')
    parser_logs.add_argument('-b', '--begin_timestamp', help='start time of logs')
    parser_logs.add_argument('-e', '--end_timestamp', help='end time of logs')
    parser_logs.add_argument('-p', '--page', type=int, default=1, help='page number, default 1')
    parser_logs.add_argument('-s', '--page_size', type=int, default=10, help='page size, default 10')
    parser_logs.set_defaults(func=get_logs)

    # get jobs
    parser_get_jobs = subparsers.add_parser('get-jobs')
    parser_get_jobs.add_argument('policy_id', type=int, help='policy id')
    parser_get_jobs.add_argument('-n', '--num', type=int, help='number of jobs')
    parser_get_jobs.add_argument('-s', '--start_time', help='start time of jobs')
    parser_get_jobs.add_argument('-e', '--end_time', help='end time of jobs')
    parser_get_jobs.add_argument('-r', '--repository', help='repo name')
    parser_get_jobs.add_argument('-t', '--status', help='job status')
    parser_get_jobs.add_argument('-p', '--page', type=int, default=1, help='page number, default 1')
    parser_get_jobs.add_argument('-z', '--page_size', type=int, default=10, help='page size, default 10')
    parser_get_jobs.set_defaults(func=get_jobs)

    # delete job
    parser_delete_job = subparsers.add_parser('delete-job')
    parser_delete_job.add_argument('id', type=long, help='job id')
    parser_delete_job.set_defaults(func=delete_job)

    # get job logs
    parser_get_job_logs = subparsers.add_parser('get-job-logs')
    parser_get_job_logs.add_argument('id', type=long, help='job id')
    parser_get_job_logs.set_defaults(func=get_job_logs)

    # get policies
    parser_get_policies = subparsers.add_parser('get-policies')
    parser_get_policies.add_argument('-n', '--name', help='policy name')
    parser_get_policies.add_argument('-i', '--project_id', type=long, help='project id')
    parser_get_policies.set_defaults(func=get_policies)

    # create policy
    parser_create_policy = subparsers.add_parser('create-policy')
    parser_create_policy.add_argument('project_id', type=long, help='project id')
    parser_create_policy.add_argument('target_id', type=int, help='target id')
    parser_create_policy.add_argument('name', help='policy name')
    parser_create_policy.add_argument('enabled', type=int, default=0, choices=[0, 1], help='0-disable, 1-enable')
    parser_create_policy.set_defaults(func=create_policy)

    # get policy
    parser_get_policy = subparsers.add_parser('get-policy')
    parser_get_policy.add_argument('id', type=long, help='policy id')
    parser_get_policy.set_defaults(func=get_policy)

    # update policy
    parser_update_policy = subparsers.add_parser('update-policy')
    parser_update_policy.add_argument('id', type=long, help='policy id')
    parser_update_policy.add_argument('-i', '--target_id', type=long, help='target id')
    parser_update_policy.add_argument('-n', '--name', help='policy name')
    parser_update_policy.add_argument('-e', '--enabled', type=int, default=0, choices=[0, 1],
                                      help='0-disable, 1-enable')
    parser_update_policy.add_argument('-d', '--description', help='policy description')
    parser_update_policy.add_argument('-c', '--cron_str', help='policy cron str,')
    parser_update_policy.set_defaults(func=update_policy)

    # enable policy
    parser_enable_policy = subparsers.add_parser('enable-policy')
    parser_enable_policy.add_argument('id', type=long, help='policy id')
    parser_enable_policy.add_argument('enabled', type=int, choices=[0, 1], help='0-disable, 1-enable')
    parser_enable_policy.set_defaults(func=enable_policy)

    # get targets
    parser_get_targets = subparsers.add_parser('get-targets')
    parser_get_targets.add_argument('-n', '--name', help='target name')
    parser_get_targets.set_defaults(func=get_targets)

    # create target
    parser_create_target = subparsers.add_parser('create-target')
    parser_create_target.add_argument('endpoint', help='target url')
    parser_create_target.add_argument('name', help='target name')
    parser_create_target.add_argument('username', help='user name')
    parser_create_target.add_argument('password', help='password of the user')
    parser_create_target.set_defaults(func=create_target)

    # test target
    parser_test_target = subparsers.add_parser('test-target')
    parser_test_target.add_argument('endpoint', help='target url')
    parser_test_target.add_argument('username', help='user name')
    parser_test_target.add_argument('password', help='password of the user')
    parser_test_target.set_defaults(func=test_target)

    # ping target
    parser_ping_target = subparsers.add_parser('ping-target')
    parser_ping_target.add_argument('id', type=long, help='target id')
    parser_ping_target.set_defaults(func=ping_target)

    # delete target
    parser_delete_target = subparsers.add_parser('delete-target')
    parser_delete_target.add_argument('id', type=long, help='target id')
    parser_delete_target.set_defaults(func=delete_target)

    # get target
    parser_get_target = subparsers.add_parser('get-target')
    parser_get_target.add_argument('id', type=long, help='target id')
    parser_get_target.set_defaults(func=get_target)

    # update target
    parser_update_target = subparsers.add_parser('update-target')
    parser_update_target.add_argument('id', type=long, help='target id')
    parser_update_target.add_argument('-e', '--endpoint', help='target url')
    parser_update_target.add_argument('-n', '--name', help='target name')
    parser_update_target.add_argument('-u', '--username', help='target user name')
    parser_update_target.add_argument('-p', '--password', help='target user password')
    parser_update_target.set_defaults(func=update_target)

    # get target policies
    parser_get_target_pol = subparsers.add_parser('get-target-policies')
    parser_get_target_pol.add_argument('id', type=long, help='target id')
    parser_get_target_pol.set_defaults(func=get_target_policies)

    # sync registry
    parser_sync_registry = subparsers.add_parser('sync-registry')
    parser_sync_registry.set_defaults(func=sync_registry)

    # get system info
    parser_get_system_info = subparsers.add_parser('get-system-info')
    parser_get_system_info.set_defaults(func=get_system_info)

    # get system volumes
    parser_get_system_volumes = subparsers.add_parser('get-system-volumes')
    parser_get_system_volumes.set_defaults(func=get_system_volumes)

    # get cert
    parser_get_cert = subparsers.add_parser('get-cert')
    parser_get_cert.set_defaults(func=get_cert)

    # ping ldap
    parser_ping_ldap = subparsers.add_parser('ping-ldap')
    parser_ping_ldap.add_argument('-w', '--ldap_url', help='ldap server url')
    parser_ping_ldap.add_argument('-d', '--ldap_search_dn', help='ldap search dn')
    parser_ping_ldap.add_argument('-p', '--ldap_search_password', help='ldap search password')
    parser_ping_ldap.add_argument('-b', '--ldap_base_dn', help='ldap base dn')
    parser_ping_ldap.add_argument('-f', '--ldap_filter', help='ldap filter')
    parser_ping_ldap.add_argument('-i', '--ldap_uid', help='ldap uid, e.g. cn, uid, sAMAccountName')
    parser_ping_ldap.add_argument('-s', '--ldap_scope', type=int, choices=[1, 2, 3], default=3,
                                  help='1-LDAP_SCOPE_BASE, 2-LDAP_SCOPE_ONELEVEL, 3-LDAP_SCOPE_SUBTREE')
    parser_ping_ldap.add_argument('-t', '--ldap_connection_timeout', type=int, default=5, help='ldap connect timeout')
    parser_ping_ldap.set_defaults(func=ping_ldap)

    # search ldap users
    parser_search_ldap_users = subparsers.add_parser('search-ldap-users')
    parser_search_ldap_users.add_argument('-u', '--username', help='ldap user name')
    parser_search_ldap_users.add_argument('-w', '--ldap_url', help='ldap server url')
    parser_search_ldap_users.add_argument('-d', '--ldap_search_dn', help='ldap search dn')
    parser_search_ldap_users.add_argument('-p', '--ldap_search_password', help='ldap search password')
    parser_search_ldap_users.add_argument('-b', '--ldap_base_dn', help='ldap base dn')
    parser_search_ldap_users.add_argument('-f', '--ldap_filter', help='ldap filter')
    parser_search_ldap_users.add_argument('-i', '--ldap_uid', help='ldap uid, e.g. cn, uid, sAMAccountName')
    parser_search_ldap_users.add_argument('-s', '--ldap_scope', type=int, choices=[1, 2, 3], default=3,
                                          help='1-LDAP_SCOPE_BASE, 2-LDAP_SCOPE_ONELEVEL, 3-LDAP_SCOPE_SUBTREE')
    parser_search_ldap_users.add_argument('-t', '--ldap_connection_timeout', type=int, default=5,
                                          help='ldap connect timeout')
    parser_search_ldap_users.set_defaults(func=search_ldap_users)

    # import ldap users
    parser_import_ldap_users = subparsers.add_parser('import-ldap-users')
    parser_import_ldap_users.add_argument('ldap_uid_list', help='ldap users\' uid list, e.g. user1,user2...')
    parser_import_ldap_users.set_defaults(func=import_ldap_users)

    # get config
    parser_get_config = subparsers.add_parser('get-config')
    parser_get_config.set_defaults(func=get_config)

    # update config
    parser_update_config = subparsers.add_parser('update-config')
    parser_update_config.add_argument('-a', '--auth_mode', help='auth mode')
    parser_update_config.add_argument('-m', '--email_from', help='email from')
    parser_update_config.add_argument('-z', '--email_host', help='email host')
    parser_update_config.add_argument('-y', '--email_identity', help='email identity')
    parser_update_config.add_argument('-x', '--email_password', help='email password')
    parser_update_config.add_argument('-o', '--email_port', help='email port')
    parser_update_config.add_argument('-l', '--email_ssl', help='email ssl')
    parser_update_config.add_argument('-u', '--email_username', help='email user name')
    parser_update_config.add_argument('-w', '--ldap_url', help='ldap server url')
    parser_update_config.add_argument('-d', '--ldap_search_dn', help='ldap search dn')
    parser_update_config.add_argument('-p', '--ldap_search_password', help='ldap search password')
    parser_update_config.add_argument('-b', '--ldap_base_dn', help='ldap base dn')
    parser_update_config.add_argument('-f', '--ldap_filter', help='ldap filter')
    parser_update_config.add_argument('-i', '--ldap_uid', help='ldap uid, e.g. cn, uid, sAMAccountName')
    parser_update_config.add_argument('-s', '--ldap_scope', type=int, choices=[1, 2, 3],
                                      help='1-LDAP_SCOPE_BASE, 2-LDAP_SCOPE_ONELEVEL, 3-LDAP_SCOPE_SUBTREE')
    parser_update_config.add_argument('-t', '--ldap_connection_timeout', type=int,
                                      help='ldap connect timeout')
    parser_update_config.add_argument('-r', '--project_creation_restriction', help='project create restriction')
    parser_update_config.add_argument('-e', '--self_registration', help='self registration')
    parser_update_config.add_argument('-v', '--verify_remote_cert', help='verify remote cert')
    parser_update_config.set_defaults(func=update_config)

    # reset config
    parser_reset_config = subparsers.add_parser('reset-config')
    parser_reset_config.set_defaults(func=reset_config)

    # ping email
    parser_ping_email = subparsers.add_parser('ping-email')
    parser_ping_email.add_argument('-o', '--email_host', help='email host')
    parser_ping_email.add_argument('-p', '--email_port', type=int, help='email port')
    parser_ping_email.add_argument('-u', '--email_username', help='email user name')
    parser_ping_email.add_argument('-w', '--email_password', help='email user password')
    parser_ping_email.add_argument('-s', '--email_ssl', type=bool, help='email ssl')
    parser_ping_email.add_argument('-i', '--email_identity', help='email identity')
    parser_ping_email.set_defaults(func=ping_email)

    args = parser.parse_args(argv)
    cli = HarborCli()
    args.func(args, cli)


def search(args, cli):
    print cli.sdk.search(args.q)


def get_projects(args, cli):
    print cli.sdk.get_projects(args.name, args.public, args.owner, args.page, args.page_size)


def check_project_exist(args, cli):
    print cli.sdk.check_project_exist(args.project_name)


def create_project(args, cli):
    print cli.sdk.create_project(args.project_name, args.public, args.enable_content_trust,
                                 args.prevent_vulnerable_images_from_running,
                                 args.prevent_vulnerable_images_from_running_severity,
                                 args.automatically_scan_images_on_push)


def delete_project(args, cli):
    print cli.sdk.delete_project(args.project_id)


def get_project(args, cli):
    print cli.sdk.get_project_info(args.project_id)


def toggle_project(args, cli):
    print cli.sdk.set_project_publicity(args.project_id, args.public)


def get_project_logs(args, cli):
    print cli.sdk.get_project_logs(args.project_id, args.username, args.repository, args.tag, args.operation,
                                   args.begin_timestamp,
                                   args.end_timestamp, args.page, args.page_size)


def get_project_members(args, cli):
    print cli.sdk.get_project_members(args.project_id)


def create_project_member(args, cli):
    roles = [0]
    roles[0] = args.role
    print cli.sdk.add_user_project_role(args.project_id, roles, args.username)


def get_project_role(args, cli):
    print cli.sdk.get_user_project_role(args.project_id, args.user_id)


def update_project_role(args, cli):
    roles = [0]
    roles[0] = args.role
    print cli.sdk.update_user_project_role(args.project_id, args.user_id, roles, args.username)


def delete_project_member(args, cli):
    print cli.sdk.delete_user_project_role(args.project_id, args.user_id)


def statistics(args, cli):
    print cli.sdk.get_statistics()


def get_users(args, cli):
    print cli.sdk.get_users(args.username, args.email, args.page, args.page_size)


def create_user(args, cli):
    print cli.sdk.create_user(args.username, args.email, args.password, args.realname, args.comment)


def get_current_user(args, cli):
    print cli.sdk.get_current_user()


def get_user(args, cli):
    print cli.sdk.get_user(args.user_id)


def update_user(args, cli):
    print cli.sdk.update_user_profile(args.user_id, args.email, args.realname, args.comment)


def delete_user(args, cli):
    print cli.sdk.delete_user(args.user_id)


def update_password(args, cli):
    print cli.sdk.change_password(args.user_id, args.old_password, args.new_password)


def toggle_admin(args, cli):
    print cli.sdk.toggle_admin(args.user_id, args.has_admin_role)


def get_repos(args, cli):
    print cli.sdk.get_repositories(args.project_id, args.repo_name, args.page, args.page_size)


def delete_repo(args, cli):
    print cli.sdk.delete_repository(args.repo_name)


def get_repo_tag(args, cli):
    print cli.sdk.get_repository_tag(args.repo_name, args.tag)


def delete_repo_tag(args, cli):
    print cli.sdk.delete_repository_tag(args.repo_name, args.tag)


def get_repo_tags(args, cli):
    print cli.sdk.get_repository_tags(args.repo_name)


def get_repo_manifest(args, cli):
    print cli.sdk.get_repository_manifest(args.repo_name, args.tag, args.version)


def create_scan_job(args, cli):
    print cli.sdk.create_scan_job(args.repo_name, args.tag)


def get_scan_detail(args, cli):
    print cli.sdk.get_scan_detail(args.repo_name, args.tag)


def get_repo_signatures(args, cli):
    print cli.sdk.get_repository_signatures(args.repo_name)


def get_top_repos(args, cli):
    print cli.sdk.get_top_repositories(args.count)


def get_logs(args, cli):
    print cli.sdk.get_logs(args.username, args.repository, args.tag, args.operation, args.begin_timestamp,
                           args.end_timestamp, args.page, args.page_size)


def get_jobs(args, cli):
    print cli.sdk.get_jobs(args.policy_id, args.num, args.start_time, args.end_time, args.repository, args.status,
                           args.page, args.page_size)


def delete_job(args, cli):
    print cli.sdk.delete_job(args.id)


def get_job_logs(args, cli):
    print cli.sdk.get_job_logs(args.id)


def get_policies(args, cli):
    print cli.sdk.get_policies(args.name, args.project_id)


def create_policy(args, cli):
    print cli.sdk.create_policy(args.project_id, args.target_id, args.name, args.enabled)


def get_policy(args, cli):
    print cli.sdk.get_policy(args.id)


def update_policy(args, cli):
    print cli.sdk.update_policy(args.id, args.target_id, args.name, args.enabled, args.description, args.cron_str)


def enable_policy(args, cli):
    print cli.sdk.update_policy_enablement(args.id, args.enabled)


def get_targets(args, cli):
    print cli.sdk.get_replication_targets(args.name)


def create_target(args, cli):
    print cli.sdk.create_replication_target(args.endpoint, args.name, args.username, args.password)


def test_target(args, cli):
    print cli.sdk.ping_replication_target(args.endpoint, args.username, args.password)


def ping_target(args, cli):
    print cli.sdk.ping_replication_target_with_id(args.id)


def delete_target(args, cli):
    print cli.sdk.delete_replication_target(args.id)


def get_target(args, cli):
    print cli.sdk.get_replication_target(args.id)


def update_target(args, cli):
    print cli.sdk.update_replication_target(args.id, args.endpoint, args.name, args.username, args.password)


def get_target_policies(args, cli):
    print cli.sdk.get_replication_target_policies(args.id)


def sync_registry(args, cli):
    print cli.sdk.sync_registry()


def get_system_info(args, cli):
    print cli.sdk.get_systeminfo()


def get_system_volumes(args, cli):
    print cli.sdk.get_systeminfo_volumes()


def get_cert(args, cli):
    print cli.sdk.get_systeminfo_cert()


def ping_ldap(args, cli):
    print cli.sdk.ping_ldap(args.ldap_url, args.ldap_search_dn, args.ldap_search_password, args.ldap_base_dn,
                            args.ldap_filter,
                            args.ldap_uid, args.ldap_scope, args.ldap_connection_timeout)


def search_ldap_users(args, cli):
    print cli.sdk.search_ldap_users(args.username, args.ldap_url, args.ldap_search_dn, args.ldap_search_password,
                                    args.ldap_base_dn, args.ldap_filter,
                                    args.ldap_uid, args.ldap_scope, args.ldap_connection_timeout)


def import_ldap_users(args, cli):
    print cli.sdk.import_ldap_users(args.ldap_uid_list.split(','))


def get_config(args, cli):
    print cli.sdk.get_configurations()


def update_config(args, cli):
    print cli.sdk.update_configurations(args.auth_mode, args.email_from, args.email_host, args.email_identity,
                                        args.email_password,
                                        args.email_port, args.email_ssl, args.email_username, args.ldap_url,
                                        args.ldap_search_dn, args.ldap_search_password, args.ldap_base_dn,
                                        args.ldap_filter,
                                        args.ldap_uid, args.ldap_scope, args.ldap_connection_timeout,
                                        args.project_creation_restriction, args.self_registration,
                                        args.verify_remote_cert)


def reset_config(args, cli):
    print cli.sdk.reset_configurations()


def ping_email(args, cli):
    print cli.sdk.ping_email(args.email_host, args.email_port, args.email_username, args.email_password, args.email_ssl,
                             args.email_identity)


if __name__ == "__main__":
    main()
