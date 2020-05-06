import { InjectionToken } from '@angular/core';

export let SERVICE_CONFIG = new InjectionToken("service.config");
export interface IServiceConfig {
    baseEndpoint?: string;
    /**
     * The base endpoint of service used to retrieve the system configuration information.
     * The configurations may include but not limit:
     *   Notary configurations
     *   Registry configuration
     *   Volume information
     *
     * * {string}
     * @memberOf IServiceConfig
     */
    systemInfoEndpoint?: string;

    /**
     * The base endpoint of the service used to handle the repositories of registry and/or tags of repository.
     * The endpoints of repository or tag(s) will be built based on this endpoint.
     * E.g:
     *   If the base endpoint is CURRENT_BASE_HREF + '/repositories',
     *   the repository endpoint will be CURRENT_BASE_HREF + '/repositories/:repo_id',
     *   the tag(s) endpoint will be CURRENT_BASE_HREF + '/repositories/:repo_id/tags[/:tag_id]'.
     *
     *
     * * {string}
     * @memberOf IServiceConfig
     */
    repositoryBaseEndpoint?: string;

    /**
     * The base endpoint of the service used to handle the recent access logs.
     *
     * * {string}
     * @memberOf IServiceConfig
     */
    logBaseEndpoint?: string;

    /**
     * The base endpoint of the service used to handle the registry targets.
     * Registry target related endpoints will be built based on this endpoint.
     * E.g:
     *   If the base endpoint is CURRENT_BASE_HREF + '/endpoints',
     *   the endpoint for registry target will be CURRENT_BASE_HREF + '/endpoints/:endpoint_id',
     *   the endpoint for pinging registry target will be CURRENT_BASE_HREF + '/endpoints/:endpoint_id/ping'.
     *
     * * {string}
     * @memberOf IServiceConfig
     */
    targetBaseEndpoint?: string;

    /**
     * The base endpoint of the service used to handle the replications.
     */
    replicationBaseEndpoint?: string;

    /**
     * The base endpoint of the service used to handle the replication rules.
     * Replication rule related endpoints will be built based on this endpoint.
     * E.g:
     *   If the base endpoint is CURRENT_BASE_HREF + '/replication/rules',
     *   the endpoint for rule will be CURRENT_BASE_HREF + '/replication/rules/:rule_id'.
     *
     * * {string}
     * @memberOf IServiceConfig
     */
    replicationRuleEndpoint?: string;

    /**
     * The base endpoint of the service used to handle vulnerability scanning.
     *
     * * {string}
     * @memberOf IServiceConfig
     */
    vulnerabilityScanningBaseEndpoint?: string;

    /**
     * The base endpoint of the service used to handle project policy.
     *
     * * {string}
     * @memberOf IServiceConfig
     */
    projectPolicyEndpoint?: string;

    /**
     * The base endpoint of service used to handle projects
     * * {string}
     * @memberOf IServiceConfig
     */
    projectBaseEndpoint?: string;

    /**
     * To determine whether or not to enable the i18 multiple languages supporting.
     *
     * * {boolean}
     * @memberOf IServiceConfig
     */
    enablei18Support?: boolean;

    /**
     * The cookie key used to store the current used language preference.
     *
     * * {string}
     * @memberOf IServiceConfig
     */
    langCookieKey?: string;

    /**
     * Declare what languages are supported.
     *
     * * {string[]}
     * @memberOf IServiceConfig
     */
    supportedLangs?: string[];

    /**
     * Define the default language the translate service uses.
     *
     * * {string}
     * @memberOf I18nConfig
     */
    defaultLang?: string;

    /**
     * To determine which loader will be used to load the required lang messages.
     * Support two loaders:
     *   One is 'http', use async http to load json files with the specified url/path.
     *   Another is 'local', use local json variable to store the lang message.
     *
     * * {string}
     * @memberOf IServiceConfig
     */
    langMessageLoader?: string;

    /**
     * Define the basic url/path prefix for the loader to find the json files if the 'langMessageLoader' is 'http'.
     * For example, 'src/i18n/langs'.
     *
     * * {string}
     * @memberOf IServiceConfig
     */
    langMessagePathForHttpLoader?: string;

    /**
     * Define the suffix of the json file names without lang name if 'langMessageLoader' is 'http'.
     * For example, '-lang.json' is suffix of message file 'en-us-lang.json'.
     *
     * * {string}
     * @memberOf IServiceConfig
     */
    langMessageFileSuffixForHttpLoader?: string;

    /**
     * If set 'local' loader in configuration property 'langMessageLoader' to load the i18n messages,
     * this property must be defined to tell local JSON loader where to get the related messages.
     * E.g:
     *   If declare the following messages storage variables,
     *
     *   export const EN_US_LANG: any = {
     *       "APP_TITLE": {
     *           "VMW_HARBOR": "VMware Harbor",
     *           "HARBOR": "Harbor"
     *       }
     *   }
     *
     *   export const ZH_CN_LANG: any = {
     *       "APP_TITLE": {
     *           "VMW_HARBOR": "VMware Harbor中文版",
     *           "HARBOR": "Harbor"
     *       }
     *   }
     *
     *   then this property should be set to:
     *   {
     *       "en-us": EN_US_LANG,
     *       "zh-cn": ZH_CN_LANG
     *   };
     *
     *
     * * {{ [key: string]: any }}
     * @memberOf IServiceConfig
     */
    localI18nMessageVariableMap?: { [key: string]: any };

    /**
     * The base endpoint of configuration service.
     *
     * * {string}
     * @memberOf IServiceConfig
     */
    configurationEndpoint?: string;

    /**
     * The base endpoint of scan job service.
     *
     * * {string}
     * @memberof IServiceConfig
     */
    scanJobEndpoint?: string;

    /**
     * The base endpoint of the service used to handle the labels.
     * labels related endpoints will be built based on this endpoint.
     * E.g:
     *   If the base endpoint is CURRENT_BASE_HREF + '/labels',
     *   the label endpoint  will be CURRENT_BASE_HREF + '/labels/:id'.
     *
     * * {string}
     * @memberOf IServiceConfig
     */
    labelEndpoint?: string;

    /**
     * The base endpoint of the service used to handle the helm chart.
     * helm charts related endpoints will be built based on this endpoint.
     * E.g:
     *   If the base endpoint is CURRENT_BASE_HREF + '/helmcharts',
     *   the helm chart endpoint  will be CURRENT_BASE_HREF + '/helmcharts/:id'.
     *
     * * {string}
     * @memberOf IServiceConfig
     */
    helmChartEndpoint?: string;

    /**
     * The base endpoint of the chart download url
     * * {string}
     */
    downloadChartEndpoint?: string;

    helmChartLabelEndpoint?: string;

    gcEndpoint?: string;

    ScanAllEndpoint?: string;

    quotaUrl?: string;
}
