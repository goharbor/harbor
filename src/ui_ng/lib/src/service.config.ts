import { OpaqueToken } from '@angular/core';

export let SERVICE_CONFIG = new OpaqueToken("service.config");

export interface IServiceConfig {
    systemInfoEndpoint?: string;

    /**
     * The base endpoint of the service used to handle the repositories of registry and/or tags of repository.
     * The endpoints of repository or tag(s) will be built based on this endpoint.
     * E.g:
     *   If the base endpoint is '/api/repositories',
     *   the repository endpoint will be '/api/repositories/:repo_id',
     *   the tag(s) endpoint will be '/api/repositories/:repo_id/tags[/:tag_id]'.
     * 
     * 
     * @type {string}
     * @memberOf IServiceConfig
     */
    repositoryBaseEndpoint?: string;

    /**
     * The base endpoint of the service used to handle the recent access logs.
     * 
     * @type {string}
     * @memberOf IServiceConfig
     */
    logBaseEndpoint?: string;

    /**
     * The base endpoint of the service used to handle the registry targets.
     * Registry target related endpoints will be built based on this endpoint.
     * E.g:
     *   If the base endpoint is '/api/endpoints',
     *   the endpoint for registry target will be '/api/endpoints/:endpoint_id',
     *   the endpoint for pinging registry target will be '/api/endpoints/:endpoint_id/ping'.
     * 
     * @type {string}
     * @memberOf IServiceConfig
     */
    targetBaseEndpoint?: string;

    /**
     * The base endpoint of the service used to handle the replication rules.
     * Replication rule related endpoints will be built based on this endpoint.
     * E.g:
     *   If the base endpoint is '/api/replication/rules',
     *   the endpoint for rule will be '/api/replication/rules/:rule_id'.
     * 
     * @type {string}
     * @memberOf IServiceConfig
     */
    replicationRuleEndpoint?: string;


    /**
     * The base endpoint of the service used to handle the replication jobs.
     * 
     * 
     * @type {string}
     * @memberOf IServiceConfig
     */
    replicationJobEndpoint?: string;

    /**
     * The cookie key used to store the current used language preference.
     * 
     * @type {string}
     * @memberOf IServiceConfig
     */
    langCookieKey?: string,

    /**
     * Declare what languages are supported.
     * 
     * @type {string[]}
     * @memberOf IServiceConfig
     */
    supportedLangs?: string[],

    /**
     * To determine whether to not enable the i18 multiple languages supporting.
     * 
     * @type {boolean}
     * @memberOf IServiceConfig
     */
    enablei18Support?: boolean
}