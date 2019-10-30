# Harbor UI library
**NOTES: Odd version is development version and may be not stable. Even version is release version which should be stable.**
![Harbor UI Library](../../../docs/img/harbor_logo.png)

Wrap the following Harbor UI components into a sharable library and published as npm package for other third-party applications to import and reuse.

* Repository and tag management view
* Replication rules and jobs management view
* Replication endpoints management view
* Access log list view
* Vulnerability scanning result bar chart and list view (Embedded in tag management view)
* Registry(Harbor) related configuration options

The Harbor UI library is built on **[Angular ](https://angular.io/)** 6.x and **[Clarity ](https://vmware.github.io/clarity/)** 0.12.x .

The library is published to the public npm repository with name **[@harbor/ui](https://www.npmjs.com/package/@harbor/ui)**.

## Build & Test
Build library with command:
```
npm run build
```

Execute the testing specs with command:
```
npm run test
```

Install the package
```
npm install @harbor/ui[@version]
```

## Usage
**Add dependency to application**

Execute install command to add dependency to package.json
```
npm install @harbor/ui

//OR
npm install @harbor/ui@1.0.0
```
The latest version of the library will be installed.

**Import the library module into the root Angular module**
```
import { HarborLibraryModule } from '@harbor/ui';

@NgModule({
    declarations: [...],
    imports: [
        HarborLibraryModule.forRoot()
    ],
    providers: [...],
    bootstrap: [...]
})
export class AppModule {
}
```
If no parameters are passed to **'forRoot'**, the module will be initialized with default configurations. If re-configuration required, please refer the **'Configurations'** parts.

**Enable components via tags**

* **Registry log view**

Use **withTitle** to set whether self-contained a header with title or not. Default is **false**, that means no header is existing.
```
<hbr-log [withTitle]="..."></hbr-log>
```

* **Replication Management View**

Support two different display scope mode: under specific project or whole system. 

If **projectId** is set to the id of specified project, then only show the replication rules bound with the project. Otherwise, show all the rules of the whole system.
On specific project mode, without need projectId, but also need to provide projectName for display.

**withReplicationJob** is used to determine whether or not show the replication jobs which are relevant with the selected replication rule.

**isSystemAdmin** is for judgment if user has administrator privilege, if true, user can do the create/edit/delete/replicate actions.

```
<hbr-replication [projectId]="..." [projectName]="..." [withReplicationJob]='...' [isSystemAdmin]="..."></hbr-replication>
```

* **Endpoint Management View**
```
//No @Input properties

<hbr-endpoint></hbr-endpoint>
```

* **Repository and Tag Management View**

The `hbr-repository-stackview` directive is deprecated. Using `hbr-repository-listview` and `hbr-repository` instead. You should define two routers one for render 
`hbr-repository-listview` the other is for `hbr-repository`. `hbr-repository-listview` will output an event, you need catch this event and redirect to related
page contains `hbr-repository`.

**hbr-repository-listview Directive**

**projectId** is used to specify which projects the repositories are from.

**projectName** is used to generate the related commands for pushing images.

**hasSignedIn** is a user session related property to determined whether a valid user signed in session existing. This component supports anonymous user.

**hasProjectAdminRole** is a user session related property to determined whether the current user has project administrator role. Some action menus might be disabled based on this property.

**repoClickEvent** is an @output event emitter for you to catch the repository click events.


```
<hbr-repository-listview [projectId]="" [projectName]="" [hasSignedIn]="" [hasProjectAdminRole]="" 
(repoClickEvent)="watchRepoClickEvent($event)"></hbr-repository-listview>

...

watchRepoClickEvent(repo: RepositoryItem): void {
    //Process repo
    ...
}
```


**hbr-repository-gridview Directive**

**projectId** is used to specify which projects the repositories are from.

**projectName** is used to generate the related commands for pushing images.

**hasSignedIn** is a user session related property to determined whether a valid user signed in session existing. This component supports anonymous user.

**hasProjectAdminRole** is a user session related property to determined whether the current user has project administrator role. Some action menus might be disabled based on this property.

**repoClickEvent** is an @output event emitter for you to catch the repository click events.

**repoProvisionEvent** is an @output event emitter for you to catch the deploy button click event.

**addInfoEvent** is an @output event emitter for you to catch the add additional info button event.

  @Output() repoProvisionEvent = new EventEmitter<RepositoryItem>();
  @Output() addInfoEvent = new EventEmitter<RepositoryItem>();


**hbr-repository Directive**

**projectId** is used to specify which projects the repositories are from.

**repoName** is used to generate the related commands for pushing images.

**hasSignedIn** is a user session related property to determined whether a valid user signed in session existing. This component supports anonymous user.

**hasProjectAdminRole** is a user session related property to determined whether the current user has project administrator role. Some action menus might be disabled based on this property.

**withNotary** is Notary installed

**tagClickEvent** is an @output event emitter for you to catch the tag click events.

**goBackClickEvent** is an @output event emitter for you to catch the go back events.

```
<hbr-repository [projectId]="" [repoName]="" [hasSignedIn]="" [hasProjectAdminRole]="" [withNotary]=""
(tagClickEvent)="watchTagClickEvt($event)" (backEvt)="watchGoBackEvt($event)" ></hbr-repository>

watchTagClickEvt(tagEvt: TagClickEvent): void {
    ...
}

watchGoBackEvt(projectId: string): void {
    ...
}
```
<hbr-repository-gridview [projectId]="" [projectName]="" [hasSignedIn]="" [hasProjectAdminRole]="" 
(repoClickEvent)="watchRepoClickEvent($event)"
(repoProvisionEvent)="watchRepoProvisionEvent($event)"
(addInfoEvent)="watchAddInfoEvent($event)"></hbr-repository-gridview>

...


watchRepoClickEvent(repo: RepositoryItem): void {
    //Process repo
    ...
}

watchRepoProvisionEvent(repo: RepositoryItem): void {
    //Process repo
    ...
}

watchAddInfoEvent(repo: RepositoryItem): void {
    //Process repo
    ...
}
```


**hbr-repository Directive**

**projectId** is used to specify which projects the repositories are from.

**repoName** is used to generate the related commands for pushing images.

**hasSignedIn** is a user session related property to determined whether a valid user signed in session existing. This component supports anonymous user.

**hasProjectAdminRole** is a user session related property to determined whether the current user has project administrator role. Some action menus might be disabled based on this property.

**tagClickEvent** is an @output event emitter for you to catch the tag click events.

**goBackClickEvent** is an @output event emitter for you to catch the go back events.

```
<hbr-repository [projectId]="" [repoName]="" [hasSignedIn]="" [hasProjectAdminRole]="" [withNotary]=""
(tagClickEvent)="watchTagClickEvt($event)" (backEvt)="watchGoBackEvt($event)" ></hbr-repository>

watchTagClickEvt(tagEvt: TagClickEvent): void {
    ...
}

watchGoBackEvt(projectId: string): void {
    ...
}
```

* **Tag detail view**

This view is linked by the repository stack view only when the Clair is enabled in Harbor.

**tagId** is an @Input property and used to specify the tag of which details are displayed.

**repositoryId** is an @Input property and used to specified the repository to which the tag is belonged.

**backEvt** is an @Output event emitter and used to distribute the click event of the back arrow in the detail page. 

```
<hbr-tag-detail (backEvt)="goBack($event)" [tagId]="..." [repositoryId]="..."></hbr-tag-detail>
```

* **Registry related configuration**

This component provides some options for registry(Harbor) related configurations.

**hasAdminRole** is an @Input property to indicate if the current logged user has administrator role.

```
<hbr-registry-config [hasAdminRole]="***"></hbr-registry-config>
```
## Configurations
All the related configurations are defined in the **HarborModuleConfig** interface.

**1. config**
The base configuration for the module. Mainly used to define the relevant endpoints of services which are in charge of retrieving data from backend APIs. It's a 'InjectionToken' and defined by 'IServiceConfig' interface. If **config** is not set, the default value will be used.
```
export const DefaultServiceConfig: IServiceConfig = {
  systemInfoEndpoint: "/api/systeminfo",
  repositoryBaseEndpoint: "/api/repositories",
  logBaseEndpoint: "/api/logs",
  targetBaseEndpoint: "/api/registries",
  replicationRuleEndpoint: "/api/policies/replication",
  replicationBaseEndpoint: "/api/replication/executions",
  vulnerabilityScanningBaseEndpoint: "/api/repositories",
  configurationEndpoint: "/api/configurations",
  enablei18Support: false,
  defaultLang: DEFAULT_LANG, //'en-us'
  langCookieKey: DEFAULT_LANG_COOKIE_KEY, //'harbor-lang'
  supportedLangs: DEFAULT_SUPPORTING_LANGS,//['en-us','zh-cn','es-es']
  langMessageLoader: "local",
  langMessagePathForHttpLoader: "i18n/langs/",
  langMessageFileSuffixForHttpLoader: "-lang.json",
  localI18nMessageVariableMap: {}
};
```
If you want to override the related items, declare your own 'IServiceConfig' interface and define the configuration value. E.g: Override 'repositoryBaseEndpoint'
```
export const MyServiceConfig: IServiceConfig = {
    repositoryBaseEndpoint: "/api/wrap/repositories"
}

...
HarborLibraryModule.forRoot({
    config: { provide: SERVICE_CONFIG, useValue: MyServiceConfig }
})
...

```
It supports partially overriding. For the items not overridden, default values will be adopted. The items contained in **config** are:
* **systemInfoEndpoint:** The base endpoint of the service used to get the related system configurations. Default value is "/api/systeminfo".

* **repositoryBaseEndpoint:** The base endpoint of the service used to handle the repositories of registry and/or tags of repository. Default value is "/api/repositories".

* **logBaseEndpoint:** The base endpoint of the service used to handle the recent access logs. Default is "/api/logs".

* **targetBaseEndpoint:** The base endpoint of the service used to handle the registry endpoints. Default is "/api/registries".

* **replicationRuleEndpoint:** The base endpoint of the service used to handle the replication rules. Default is "/api/policies/replication".

* **replicationBaseEndpoint:** The base endpoint of the service used to handle the replication executions. Default is "/api/replication/executions".

* **vulnerabilityScanningBaseEndpoint:** The base endpoint of the service used to handle the vulnerability scanning results.Default value is "/api/repositories".

* **configurationEndpoint:** The base endpoint of the service used to configure registry related options. Default is "/api/configurations".

* **langCookieKey:** The cookie key used to store the current used language preference. Default is "harbor-lang".

* **supportedLangs:** Declare what languages are supported. Default is ['en-us', 'zh-cn', 'es-es'].

* **enablei18Support:** To determine whether or not to enable the i18 multiple languages supporting. Default is false.

* **langMessageLoader:** To determine which loader will be used to load the required lang messages. Support two loaders: One is **'http'**, use async http to load json files with the specified url/path. Another is **'local'**, use local json variable to store the lang message.

* **langMessagePathForHttpLoader:** Define the basic url/path prefix for the loader to find the json files if the 'langMessageLoader' is set to **'http'**. E.g: 'src/i18n/langs'.

* **langMessageFileSuffixForHttpLoader:** Define the suffix of the json file names without lang name if 'langMessageLoader' is set to **'http'**. For example, '-lang.json' is suffix of message file 'en-us-lang.json'.

* **localI18nMessageVariableMap:** If configuration property 'langMessageLoader' is set to **'local'** to load the i18n messages, this property must be defined to tell local JSON loader where to get the related messages. E.g: If declare the following messages storage variables,
```
        export const EN_US_LANG: any = {
            "APP_TITLE": {
                "VMW_HARBOR": "VMware Harbor",
                "HARBOR": "Harbor"
            }
        }
        
        export const ZH_CN_LANG: any = {
            "APP_TITLE": {
                "VMW_HARBOR": "VMware Harbor中文版",
                "HARBOR": "Harbor"
            }
        }
```

then this property should be set to:
```
        {
            "en-us": EN_US_LANG,
            "zh-cn": ZH_CN_LANG
        };
```

**2. errorHandler**
UI components in the library use this interface to pass the errors/warnings/infos/logs to the top component or page. The top component or page can display those information in their message panel or notification system.
If not set, the console will be used as default output approach.

```
@Injectable()
export class MyErrorHandler extends ErrorHandler {
    public error(error: any): void {
        ...
    }

    public warning(warning: any): void {
        ...
    }

    public info(info: any): void {
        ...
    }

    public log(log: any): void {
        ...
    }
}

...
HarborLibraryModule.forRoot({
    errorHandler: { provide: ErrorHandler, useClass: MyErrorHandler }
})
...

```
**3. user session**
Some components may need the user authorization and authentication information to display different views. The following way of handing user session is supported by the library.
* Use @Input properties or interface to let top component or page to pass the required user session information in.

```
//In the above repository stack view, the user session informations are passed via @input properties.
[hasSignedIn]="..." [hasProjectAdminRole]="..."
```
**4. services**
The library has its own service implementations to communicate with backend APIs and transfer data. If you want to use your own data handling logic, you can implement your own services based on the defined interfaces.

* **AccessLogService:** Define service methods to handle the access log related things.
```
@Injectable()
export class MyAccessLogService extends AccessLogService {
     /**
     * Get the audit logs for the specified project.
     * Set query parameters through 'queryParams', support:
     *  - page
     *  - pageSize
     * 
     * @param {(number | string)} projectId
     * @param {RequestQueryParams} [queryParams]
     * @returns {(Observable<AccessLog[]>)}
     * 
     * @memberOf AccessLogService
     */
    getAuditLogs(projectId: number | string, queryParams?: RequestQueryParams): Observable<AccessLog[]>  {
        ...
    }

    /**
     * Get the recent logs.
     * 
     * @param {number} lines : Specify how many lines should be returned.
     * @returns {(Observable<AccessLog[]>)}
     * 
     * @memberOf AccessLogService
     */
    getRecentLogs(lines: number): Observable<AccessLog[]>{
        ...
    }
}

...
HarborLibraryModule.forRoot({
    logService: { provide: AccessLogService, useClass: MyAccessLogService }
})
...

```

* **EndpointService:** Define the service methods to handle the endpoint related things.
```
@Injectable()
export class MyEndpointService extends EndpointService {
    /**
     * Get all the endpoints.
     * Set the argument 'endpointName' to return only the endpoints match the name pattern.
     * 
     * @param {string} [endpointName]
     * @param {RequestQueryParams} [queryParams]
     * @returns {(Observable<Endpoint[]> | Endpoint[])}
     * 
     * @memberOf EndpointService
     */
    getEndpoints(endpointName?: string, queryParams?: RequestQueryParams): Observable<Endpoint[]> {
        ...
    }

    /**
     * Get the specified endpoint.
     * 
     * @param {(number | string)} endpointId
     * @returns {(Observable<Endpoint> | Endpoint)}
     * 
     * @memberOf EndpointService
     */
    getEndpoint(endpointId: number | string): Observable<Endpoint>{
        ...
    }

    /**
     * Create new endpoint.
     * 
     * @param {Endpoint} endpoint
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf EndpointService
     */
    createEndpoint(endpoint: Endpoint): Observable<any> {
        ...
    }

    /**
     * Update the specified endpoint.
     * 
     * @param {(number | string)} endpointId
     * @param {Endpoint} endpoint
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf EndpointService
     */
    updateEndpoint(endpointId: number | string, endpoint: Endpoint): Observable<any> {
        ...
    }

    /**
     * Delete the specified endpoint.
     * 
     * @param {(number | string)} endpointId
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf EndpointService
     */
    deleteEndpoint(endpointId: number | string): Observable<any> {
        ...
    }

    /**
     * Ping the specified endpoint.
     * 
     * @param {Endpoint} endpoint
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf EndpointService
     */
    pingEndpoint(endpoint: Endpoint): Observable<any> {
        ...
    }

    /**
     * Check endpoint whether in used with specific replication rule.
     * 
     * @param {{number | string}} endpointId
     * @returns {{Observable<any> | any}}
     */
    getEndpointWithReplicationRules(endpointId: number | string): Observable<any> {
        ...
    }
}

...
HarborLibraryModule.forRoot({
    endpointService: { provide: EndpointService, useClass: MyEndpointService }
})
...

```

* **ReplicationService:** Define the service methods to handle the replication (rule and job) related things.
```
@Injectable()
export class MyReplicationService extends ReplicationService {
     /**
     * Get the replication rules.
     * Set the argument 'projectId' to limit the data scope to the specified project;
     * set the argument 'ruleName' to return the rule only match the name pattern;
     * if pagination needed, use the queryParams to add query parameters.
     * 
     * @param {(number | string)} [projectId]
     * @param {string} [ruleName]
     * @param {RequestQueryParams} [queryParams]
     * @returns {(Observable<ReplicationRule[]>)}
     * 
     * @memberOf ReplicationService
     */
    getReplicationRules(projectId?: number | string, ruleName?: string, queryParams?: RequestQueryParams): Observable<ReplicationRule[]> {
        ...
    }

    /**
     * Get the specified replication rule.
     * 
     * @param {(number | string)} ruleId
     * @returns {(Observable<ReplicationRule>)}
     * 
     * @memberOf ReplicationService
     */
    getReplicationRule(ruleId: number | string): Observable<ReplicationRule> {
        ...
    }

    /**
     * Create new replication rule.
     * 
     * @param {ReplicationRule} replicationRule
     * @returns {(Observable<any>)}
     * 
     * @memberOf ReplicationService
     */
    createReplicationRule(replicationRule: ReplicationRule): Observable<any> {
        ...
    }

    /**
     * Update the specified replication rule.
     * 
     * @param {ReplicationRule} replicationRule
     * @returns {(Observable<any>)}
     * 
     * @memberOf ReplicationService
     */
    updateReplicationRule(replicationRule: ReplicationRule): Observable<any> {
        ...
    }

    /**
     * Delete the specified replication rule.
     * 
     * @param {(number | string)} ruleId
     * @returns {(Observable<any>)}
     * 
     * @memberOf ReplicationService
     */
    deleteReplicationRule(ruleId: number | string): Observable<any> {
        ...
    }

    /**
     * Enable the specified replication rule.
     * 
     * @param {(number | string)} ruleId
     * @returns {(Observable<any>)}
     * 
     * @memberOf ReplicationService
     */
    enableReplicationRule(ruleId: number | string, enablement: number): Observable<any> {
        ...
    }

    /**
     * Disable the specified replication rule.
     * 
     * @param {(number | string)} ruleId
     * @returns {(Observable<any>)}
     * 
     * @memberOf ReplicationService
     */
    disableReplicationRule(ruleId: number | string): Observable<any> {
        ...
    }

    /**
     * Get the jobs for the specified replication rule.
     * Set query parameters through 'queryParams', support:
     *   - status
     *   - repository
     *   - startTime and endTime
     *   - page
     *   - pageSize
     * 
     * @param {(number | string)} ruleId
     * @param {RequestQueryParams} [queryParams]
     * @returns {(Observable<ReplicationJob>)}
     * 
     * @memberOf ReplicationService
     */
    getJobs(ruleId: number | string, queryParams?: RequestQueryParams): Observable<ReplicationJob[]> {
        ...
    }
}

...
HarborLibraryModule.forRoot({
    replicationService: { provide: ReplicationService, useClass: MyReplicationService }
})
...

```

* **RepositoryService:**  Define service methods for handling the repository related things.
```
@Injectable()
export class MyRepositoryService extends RepositoryService {
    /**
     * List all the repositories in the specified project.
     * Specify the 'repositoryName' to only return the repositories which match the name pattern.
     * If pagination needed, set the following parameters in queryParams:
     *   'page': current page,
     *   'page_size': page size.
     * 
     * @param {(number | string)} projectId
     * @param {string} repositoryName
     * @param {RequestQueryParams} [queryParams]
     * @returns {(Observable<Repository[]>)}
     * 
     * @memberOf RepositoryService
     */
    getRepositories(projectId: number | string, repositoryName?: string, queryParams?: RequestQueryParams): Observable<Repository[]> {
        ...
    }

    /**
     * DELETE the specified repository.
     * 
     * @param {string} repositoryName
     * @returns {(Observable<any>)}
     * 
     * @memberOf RepositoryService
     */
    deleteRepository(repositoryName: string): Observable<any> {
        ...
    }
}

...
HarborLibraryModule.forRoot({
    repositoryService: { provide: RepositoryService, useClass: MyRepositoryService }
})
...

```

```
@Injectable()
export class MyTagService extends TagService {
    /**
     * Get all the tags under the specified repository.
     * NOTES: If the Notary is enabled, the signatures should be included in the returned data.
     *  
     * @param {string} repositoryName
     * @param {RequestQueryParams} [queryParams]
     * @returns {(Observable<Tag[]>)}
     * 
     * @memberOf TagService
     */
    getTags(repositoryName: string, queryParams?: RequestQueryParams): Observable<Tag[]> {
        ...
    }

    /**
     * Delete the specified tag.
     * 
     * @param {string} repositoryName
     * @param {string} tag
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf TagService
     */
    deleteTag(repositoryName: string, tag: string): Observable<any> {
        ...
    }
}

...
HarborLibraryModule.forRoot({
    tagService: { provide: TagService, useClass: MyTagService }
})
...

```

* **ScanningResultService:** Get the vulnerabilities scanning results for the specified tag.
```
@Injectable()
/**
 * Get the vulnerabilities scanning results for the specified tag.
 * 
 * @export
 * @class ScanningResultService
 */
export class MyScanningResultService extends ScanningResultService {
    /**
     * Get the summary of vulnerability scanning result.
     * 
     * @param {string} tagId
     * @returns {(Observable<VulnerabilitySummary>)}
     * 
     * @memberOf ScanningResultService
     */
    getVulnerabilityScanningSummary(repoName: string, tagId: string, queryParams?: RequestQueryParams): Observable<VulnerabilitySummary> {
        ...
    }

    /**
     * Get the detailed vulnerabilities scanning results.
     * 
     * @param {string} tagId
     * @returns {(Observable<VulnerabilityItem[]>)}
     * 
     * @memberOf ScanningResultService
     */
    getVulnerabilityScanningResults(repoName: string, tagId: string, queryParams?: RequestQueryParams): Observable<VulnerabilityItem[]> {
        ...
    }


    /**
     * Start a new vulnerability scanning
     * 
     * @param {string} repoName
     * @param {string} tagId
     * @returns {(Observable<any>)}
     * 
     * @memberOf ScanningResultService
     */
    startVulnerabilityScanning(repoName: string, tagId: string): Observable<any> {
        ...
    }
}

...
HarborLibraryModule.forRoot({
    scanningService: { provide: ScanningResultService, useClass: MyScanningResultService }
})
...

```

* **SystemInfoService:** Get related system configurations.
```
/**
 * Get System information about current backend server.
 * @class
 */
export class MySystemInfoService extends SystemInfoService {
  /**
   *  Get global system information.
   *  @returns 
   */
   getSystemInfo(): Observable<SystemInfo> {
       ...
   }
}

...
HarborLibraryModule.forRoot({
    systemInfoService: { provide: SystemInfoService, useClass: MySystemInfoService }
})
...

```

* **ConfigurationService:** Get and save the registry related configuration options.

```
/**
 * Service used to get and save registry-related configurations.
 * 
 * @export
 * @class MyConfigurationService
 */
export class MyConfigurationService extends ConfigurationService{

    /**
     * Get configurations.
     * 
    
     * @returns {(Observable<Configuration>)}
     * 
     * @memberOf ConfigurationService
     */
    getConfigurations(): Observable<Configuration> {
        ...
    }

    /**
     * Save configurations.
     * 
    
     * @returns {(Observable<Configuration>)}
     * 
     * @memberOf ConfigurationService
     */
    saveConfigurations(changedConfigs: any | { [key: string]: any | any[] }): Observable<any> {
        ...
    }
}

...
HarborLibraryModule.forRoot({
    config.configService || { provide: ConfigurationService, useClass: ConfigurationDefaultService }
})
...
```
