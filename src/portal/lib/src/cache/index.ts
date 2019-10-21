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

import { Observable, of } from "rxjs";
import { tap, publishReplay, refCount } from "rxjs/operators";

function hashCode(str: string): string {
    let hash: number = 0;
    let chr: number;
    if (str.length === 0) {
        return hash.toString(36);
    }

    for (let i = 0; i < str.length; i++) {
        chr = str.charCodeAt(i);

        /* tslint:disable:no-bitwise */
        hash = ((hash << 5) - hash) + chr;
        hash |= 0;
        /* tslint:enable:no-bitwise */
    }

    return hash.toString(36);
}

interface IObservableCacheValue {
    response: Observable<any>;

    /**
     * created time of the cache value
     */
    created?: Date;
}

interface IObservableCacheConfig {
    /**
     * maxAge of cache in milliseconds
     */
    maxAge?: number;

    /**
     * whether should use sliding expiration on caches
     */
    slidingExpiration?: boolean;
}

const cache: Map<string, IObservableCacheValue> = new Map<string, IObservableCacheValue>();

export function CacheObservable(config: IObservableCacheConfig = {}) {
    return function (target: any, methodName: string, descriptor: PropertyDescriptor) {
        const original = descriptor.value;
        const targetName = target.constructor.name;

        (descriptor.value as any) = function (...args: Array<any>) {
            const key = hashCode(`${targetName}:${methodName}:${JSON.stringify(args)}`);

            let value = cache.get(key);
            if (value && value.created) {
                if (new Date().getTime() - new Date(value.created).getTime() > config.maxAge) {
                    cache[key] = null;
                    value = null;
                } else if (config.slidingExpiration) {
                    value.created = new Date();
                    cache.set(key, value);
                }
            }

            if (value) {
                return of(value.response);
            }

            const response$ = (original.apply(this, args) as Observable<any>).pipe(
                tap((response: Observable<any>) => {
                    cache.set(key,  {
                        response: response,
                        created: config.maxAge ? new Date() : null
                    });
                }),
                publishReplay(1),
                refCount()
            );

            return response$;
        };

        return descriptor;
    };
}

export function FlushAll() {
    cache.clear();
}
