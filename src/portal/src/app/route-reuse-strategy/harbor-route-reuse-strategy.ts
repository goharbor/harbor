// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import {
    RouteReuseStrategy,
    ActivatedRouteSnapshot,
    DetachedRouteHandle,
} from '@angular/router';

/**
 * if want to reuse a route, add  reuse: true to its routeConfig data as below:
 * data : {
 *     reuse: true,
 *     routeConfigId: 'one unique id'
 * }
 */

export enum RouteConfigId {
    REPLICATION_PAGE = 'TotalReplicationPageComponent',
    REPLICATION_TASKS_PAGE = 'ReplicationTasksComponent',
    P2P_POLICIES_PAGE = 'PolicyComponent',
    P2P_TASKS_PAGE = 'P2pTaskListComponent',
    WEBHOOK_POLICIES_PAGE = 'WebhookComponent',
    WEBHOOK_TASKS_PAGE = 'WebhookTasksComponent',
}
// should not reuse the routes that meet these RegExps
const ShouldNotReuseRouteRegExps: RegExp[] = [
    /\/harbor\/projects\/(\d+)\/repositories$/,
    /\/harbor\/projects\/(\d+)\/repositories\/(\S+)\/artifacts-tab$/,
    /\/harbor\/projects\/(\d+)\/repositories\/(\S+)\/artifacts-tab\/depth\/(\S+)$/,
];

function testRoute(url: string) {
    let flag: boolean = false;
    ShouldNotReuseRouteRegExps.forEach(item => {
        if (item.test(url)) {
            flag = true;
        }
    });
    return flag;
}

export class HarborRouteReuseStrategy implements RouteReuseStrategy {
    /**
     * 1.for each routing action, cache will be removed by default
     * 2.add the routing actions here that should keep cache
     * 3.you need to add routeConfigId: 'one unique id' to the related router configs like below:
     *  data : {
     *     reuse: true,
     *     routeConfigId: 'one unique id'
     * }
     * @param future
     * @param curr
     * @private
     */
    private shouldKeepCache(
        future: ActivatedRouteSnapshot,
        curr: ActivatedRouteSnapshot
    ) {
        if (
            future.routeConfig &&
            curr.routeConfig &&
            future.routeConfig.data &&
            curr.routeConfig.data
        ) {
            // action 1: from replication tasks list page to TotalReplicationPageComponent page
            if (
                curr.routeConfig.data.routeConfigId ===
                    RouteConfigId.REPLICATION_TASKS_PAGE &&
                future.routeConfig.data.routeConfigId ===
                    RouteConfigId.REPLICATION_PAGE
            ) {
                this.shouldDeleteCache = false;
            }
            // action 2: from preheat tasks list page to PolicyComponent page
            if (
                curr.routeConfig.data.routeConfigId ===
                    RouteConfigId.P2P_TASKS_PAGE &&
                future.routeConfig.data.routeConfigId ===
                    RouteConfigId.P2P_POLICIES_PAGE
            ) {
                this.shouldDeleteCache = false;
            }
            // action 3: from webhook tasks list page to WebhookComponent page
            if (
                curr.routeConfig.data.routeConfigId ===
                    RouteConfigId.WEBHOOK_TASKS_PAGE &&
                future.routeConfig.data.routeConfigId ===
                    RouteConfigId.WEBHOOK_POLICIES_PAGE
            ) {
                this.shouldDeleteCache = false;
            }
        }
    }

    private _cache: { [key: string]: DetachedRouteHandle } = {};

    // cache will be removed by default
    private shouldDeleteCache: boolean = true;

    shouldReuseRoute(
        future: ActivatedRouteSnapshot,
        curr: ActivatedRouteSnapshot
    ): boolean {
        this.shouldKeepCache(future, curr);
        if (
            testRoute(curr['_routerState']?.url) &&
            testRoute(future['_routerState']?.url)
        ) {
            return false;
        }
        return future.routeConfig === curr.routeConfig;
    }

    shouldAttach(route: ActivatedRouteSnapshot): boolean {
        if (this.isReuseRoute(route)) {
            if (this.shouldDeleteCache) {
                this.clearAllCache();
            }
        }
        setTimeout(() => {
            this.shouldDeleteCache = true;
        }, 0);
        return this._cache[this.getFullUrl(route)] && this.isReuseRoute(route);
    }

    retrieve(route: ActivatedRouteSnapshot): DetachedRouteHandle {
        if (this._cache[this.getFullUrl(route)] && this.isReuseRoute(route)) {
            return this._cache[this.getFullUrl(route)];
        }
        return null;
    }

    shouldDetach(route: ActivatedRouteSnapshot): boolean {
        return this.isReuseRoute(route);
    }

    store(route: ActivatedRouteSnapshot, handle: DetachedRouteHandle): void {
        // use the full urls as cache keys
        this._cache[this.getFullUrl(route)] = handle;
    }

    // full url, equals to window.location.pathName
    private getFullUrl(route: ActivatedRouteSnapshot) {
        return route['_routerState'].url;
    }

    // if this route should be reused
    private isReuseRoute(route: ActivatedRouteSnapshot): boolean {
        return (
            route &&
            route.routeConfig &&
            route.routeConfig.data &&
            route.routeConfig.data.reuse
        );
    }

    // clear cache
    private clearAllCache() {
        for (let name in this._cache) {
            if (this._cache.hasOwnProperty(name)) {
                if (this._cache[name]) {
                    if ((this._cache[name] as any).componentRef) {
                        (this._cache[name] as any).componentRef.destroy(); // manually call destroy(), to destroy component
                    }
                }
                delete this._cache[name];
            }
        }
    }
}
