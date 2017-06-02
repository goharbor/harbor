# Harbor UI library
Wrap the following Harbor UI components into a sharable library and published as npm package for other third-party applications to import and reuse.

* Repository and tag management view
* Replication rules and jobs management view
* Replication endpoints management view
* Access log list view
* Vulnerability scanning result bar chart and list view (Embedded in tag management view)

The Harbor UI library is built on **[Angular ](https://angular.io/)** 4.x and **[Clarity ](https://vmware.github.io/clarity/)** 0.9.x .

The library is published to the public npm repository with name **[harbor-ui](https://www.npmjs.com/package/harbor-ui)**.

## Build & Test
Build library with command:
```
npm run build
```

Execute the testing specs with command:
```
npm run test
```

## Usage
**Add dependency to application**

Execute install command to add dependency to package.json
```
npm install harbor-ui --save
```
The latest version of the library will be installed.

**Import the library module into the root Angular module**
```
import { HarborLibraryModule } from 'harbor-ui';

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
If no parameters are passed to **'forRoot'**, the module will be initialized with default configurations. If re-configuration required, please refer the **Configurations** parts.

## Configurations
All the related configurations are defined in the **HarborModuleConfig** interface.

**1. config**
The base configuration for the module. Mainly used to define the relevant endpoints of services which are in charge of retrieving data from backend APIs. It's a 'OpaqueToken' and defined by 'IServiceConfig' interface. If **config** is not set, the default value will be used.
```
export const DefaultServiceConfig: IServiceConfig = {
  repositoryBaseEndpoint: "/api/repositories",
  logBaseEndpoint: "/api/logs",
  targetBaseEndpoint: "/api/targets",
  replicationRuleEndpoint: "/api/policies/replication",
  replicationJobEndpoint: "/api/jobs/replication",
  langCookieKey: DEFAULT_LANG_COOKIE_KEY,
  supportedLangs: DEFAULT_SUPPORTING_LANGS,
  enablei18Support: false
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
* **repositoryBaseEndpoint:** The base endpoint of the service used to handle the repositories of registry and/or tags of repository. Default value is "/api/repositories".
* **logBaseEndpoint:** The base endpoint of the service used to handle the recent access logs. Default is "/api/logs".
* **targetBaseEndpoint:** The base endpoint of the service used to handle the registry endpoints. Default is "/api/targets".
* **replicationRuleEndpoint:** The base endpoint of the service used to handle the replication rules. Default is "/api/policies/replication".
* **replicationJobEndpoint:** The base endpoint of the service used to handle the replication jobs. Default is "/api/jobs/replication".
* **langCookieKey:** The cookie key used to store the current used language preference. Default is "harbor-lang".
* **supportedLangs:** Declare what languages are supported. Default is ['en-us', 'zh-cn', 'es-es'].
* **enablei18Support:** To determine whether to not enable the i18 multiple languages supporting. Default is false.

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
**3. user session(Ongoing/Discussing)**
Some components may need the user authorization and authentication information to display different views. There might be two alternatives to select:
* Use @Input properties or interface to let top component or page to pass the required user session information in.
* Component retrieves the required information from some API provided by top component or page when necessary. 

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
     * @abstract
     * @param {(number | string)} projectId
     * @param {RequestQueryParams} [queryParams]
     * @returns {(Observable<AccessLog[]> | Promise<AccessLog[]> | AccessLog[])}
     * 
     * @memberOf AccessLogService
     */
    getAuditLogs(projectId: number | string, queryParams?: RequestQueryParams): Observable<AccessLog[]> | Promise<AccessLog[]> | AccessLog[]{
        ...
    }

    /**
     * Get the recent logs.
     * 
     * @abstract
     * @param {number} lines : Specify how many lines should be returned.
     * @returns {(Observable<AccessLog[]> | Promise<AccessLog[]> | AccessLog[])}
     * 
     * @memberOf AccessLogService
     */
    getRecentLogs(lines: number): Observable<AccessLog[]> | Promise<AccessLog[]> | AccessLog[]{
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
     * @abstract
     * @param {string} [endpointName]
     * @param {RequestQueryParams} [queryParams]
     * @returns {(Observable<Endpoint[]> | Endpoint[])}
     * 
     * @memberOf EndpointService
     */
    getEndpoints(endpointName?: string, queryParams?: RequestQueryParams): Observable<Endpoint[]> | Promise<Endpoint[]> | Endpoint[] {
        ...
    }

    /**
     * Get the specified endpoint.
     * 
     * @abstract
     * @param {(number | string)} endpointId
     * @returns {(Observable<Endpoint> | Endpoint)}
     * 
     * @memberOf EndpointService
     */
    getEndpoint(endpointId: number | string): Observable<Endpoint> | Promise<Endpoint> | Endpoint {
        ...
    }

    /**
     * Create new endpoint.
     * 
     * @abstract
     * @param {Endpoint} endpoint
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf EndpointService
     */
    createEndpoint(endpoint: Endpoint): Observable<any> | Promise<any> | any {
        ...
    }

    /**
     * Update the specified endpoint.
     * 
     * @abstract
     * @param {(number | string)} endpointId
     * @param {Endpoint} endpoint
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf EndpointService
     */
    updateEndpoint(endpointId: number | string, endpoint: Endpoint): Observable<any> | Promise<any> | any {
        ...
    }

    /**
     * Delete the specified endpoint.
     * 
     * @abstract
     * @param {(number | string)} endpointId
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf EndpointService
     */
    deleteEndpoint(endpointId: number | string): Observable<any> | Promise<any> | any {
        ...
    }

    /**
     * Ping the specified endpoint.
     * 
     * @abstract
     * @param {Endpoint} endpoint
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf EndpointService
     */
    pingEndpoint(endpoint: Endpoint): Observable<any> | Promise<any> | any {
        ...
    }

    /**
     * Check endpoint whether in used with specific replication rule.
     * 
     * @abstract 
     * @param {{number | string}} endpointId
     * @returns {{Observable<any> | any}}
     */
    getEndpointWithReplicationRules(endpointId: number | string): Observable<any> | Promise<any> | any {
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
     * @abstract
     * @param {(number | string)} [projectId]
     * @param {string} [ruleName]
     * @param {RequestQueryParams} [queryParams]
     * @returns {(Observable<ReplicationRule[]> | Promise<ReplicationRule[]> | ReplicationRule[])}
     * 
     * @memberOf ReplicationService
     */
    getReplicationRules(projectId?: number | string, ruleName?: string, queryParams?: RequestQueryParams): Observable<ReplicationRule[]> | Promise<ReplicationRule[]> | ReplicationRule[] {
        ...
    }

    /**
     * Get the specified replication rule.
     * 
     * @abstract
     * @param {(number | string)} ruleId
     * @returns {(Observable<ReplicationRule> | Promise<ReplicationRule> | ReplicationRule)}
     * 
     * @memberOf ReplicationService
     */
    getReplicationRule(ruleId: number | string): Observable<ReplicationRule> | Promise<ReplicationRule> | ReplicationRule {
        ...
    }

    /**
     * Create new replication rule.
     * 
     * @abstract
     * @param {ReplicationRule} replicationRule
     * @returns {(Observable<any> | Promise<any> | any)}
     * 
     * @memberOf ReplicationService
     */
    createReplicationRule(replicationRule: ReplicationRule): Observable<any> | Promise<any> | any {
        ...
    }

    /**
     * Update the specified replication rule.
     * 
     * @abstract
     * @param {ReplicationRule} replicationRule
     * @returns {(Observable<any> | Promise<any> | any)}
     * 
     * @memberOf ReplicationService
     */
    updateReplicationRule(replicationRule: ReplicationRule): Observable<any> | Promise<any> | any {
        ...
    }

    /**
     * Delete the specified replication rule.
     * 
     * @abstract
     * @param {(number | string)} ruleId
     * @returns {(Observable<any> | Promise<any> | any)}
     * 
     * @memberOf ReplicationService
     */
    deleteReplicationRule(ruleId: number | string): Observable<any> | Promise<any> | any {
        ...
    }

    /**
     * Enable the specified replication rule.
     * 
     * @abstract
     * @param {(number | string)} ruleId
     * @returns {(Observable<any> | Promise<any> | any)}
     * 
     * @memberOf ReplicationService
     */
    enableReplicationRule(ruleId: number | string, enablement: number): Observable<any> | Promise<any> | any {
        ...
    }

    /**
     * Disable the specified replication rule.
     * 
     * @abstract
     * @param {(number | string)} ruleId
     * @returns {(Observable<any> | Promise<any> | any)}
     * 
     * @memberOf ReplicationService
     */
    disableReplicationRule(ruleId: number | string): Observable<any> | Promise<any> | any {
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
     * @abstract
     * @param {(number | string)} ruleId
     * @param {RequestQueryParams} [queryParams]
     * @returns {(Observable<ReplicationJob> | Promise<ReplicationJob[]> | ReplicationJob)}
     * 
     * @memberOf ReplicationService
     */
    getJobs(ruleId: number | string, queryParams?: RequestQueryParams): Observable<ReplicationJob[]> | Promise<ReplicationJob[]> | ReplicationJob[] {
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
     * @abstract
     * @param {(number | string)} projectId
     * @param {string} repositoryName
     * @param {RequestQueryParams} [queryParams]
     * @returns {(Observable<Repository[]> | Promise<Repository[]> | Repository[])}
     * 
     * @memberOf RepositoryService
     */
    getRepositories(projectId: number | string, repositoryName?: string, queryParams?: RequestQueryParams): Observable<Repository[]> | Promise<Repository[]> | Repository[] {
        ...
    }

    /**
     * DELETE the specified repository.
     * 
     * @abstract
     * @param {string} repositoryName
     * @returns {(Observable<any> | Promise<any> | any)}
     * 
     * @memberOf RepositoryService
     */
    deleteRepository(repositoryName: string): Observable<any> | Promise<any> | any {
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
     * @abstract
     * @param {string} repositoryName
     * @param {RequestQueryParams} [queryParams]
     * @returns {(Observable<Tag[]> | Promise<Tag[]> | Tag[])}
     * 
     * @memberOf TagService
     */
    getTags(repositoryName: string, queryParams?: RequestQueryParams): Observable<Tag[]> | Promise<Tag[]> | Tag[] {
        ...
    }

    /**
     * Delete the specified tag.
     * 
     * @abstract
     * @param {string} repositoryName
     * @param {string} tag
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf TagService
     */
    deleteTag(repositoryName: string, tag: string): Observable<any> | Promise<Tag> | any {
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
export class MyScanningResultService extends ScanningResultService {
    /**
     * Get the summary of vulnerability scanning result.
     * 
     * @abstract
     * @param {string} tagId
     * @returns {(Observable<ScanningResultSummary> | Promise<ScanningResultSummary> | ScanningResultSummary)}
     * 
     * @memberOf ScanningResultService
     */
    getScanningResultSummary(tagId: string): Observable<ScanningResultSummary> | Promise<ScanningResultSummary> | ScanningResultSummary {
        ...
    }

    /**
     * Get the detailed vulnerabilities scanning results.
     * 
     * @abstract
     * @param {string} tagId
     * @returns {(Observable<ScanningDetailResult[]> | Promise<ScanningDetailResult[]> | ScanningDetailResult[])}
     * 
     * @memberOf ScanningResultService
     */
    getScanningResults(tagId: string): Observable<ScanningDetailResult[]> | Promise<ScanningDetailResult[]> | ScanningDetailResult[] {
        ...
    }
}

...
HarborLibraryModule.forRoot({
    scanningService: { provide: ScanningResultService, useClass: MyScanningResultService }
})
...

```