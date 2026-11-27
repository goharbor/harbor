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
import { TestBed } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import {
    ActivatedRouteSnapshot,
    Router,
    RouterStateSnapshot,
    UrlSegment,
    UrlTree,
} from '@angular/router';
import { Observable, of, throwError } from 'rxjs';
import { ErrorHandler } from '../../../shared/units/error-handler';
import { UserPermissionService } from '../../../shared/services';
import {
    TagStrategyDefaultTabGuardService,
    TagStrategyGuardService,
} from './tag-strategy-guard.service';

describe('TagStrategyGuardService', () => {
    let permissions: Observable<boolean[]>;
    const fakeUserPermissionService = {
        hasProjectPermissions() {
            return permissions;
        },
    };
    const fakeErrorHandler = {
        error() {},
    };
    const state = {} as RouterStateSnapshot;

    function snapshot(
        path: string,
        parent: ActivatedRouteSnapshot,
        params: any = {}
    ): ActivatedRouteSnapshot {
        const url = path ? [new UrlSegment(path, {})] : [];
        const pathFromRoot = parent ? [...parent.pathFromRoot] : [];
        const route = { url, parent, params, pathFromRoot } as any;
        pathFromRoot.push(route);
        return route as ActivatedRouteSnapshot;
    }

    // /harbor/projects/1/tag-strategy
    const root = snapshot('', null);
    const harbor = snapshot('harbor', root);
    const projects = snapshot('projects', harbor);
    const project = snapshot('1', projects, { id: 1 });
    const tagStrategy = snapshot('tag-strategy', project);
    const tagFeature = snapshot('', tagStrategy);
    const defaultTab = snapshot('', tagFeature);

    let router: Router;

    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [RouterTestingModule],
            providers: [
                {
                    provide: UserPermissionService,
                    useValue: fakeUserPermissionService,
                },
                { provide: ErrorHandler, useValue: fakeErrorHandler },
            ],
        });
        router = TestBed.inject(Router);
    });

    function resolve(
        result: Observable<boolean | UrlTree>
    ): Promise<boolean | UrlTree> {
        return new Promise(resolvePromise => result.subscribe(resolvePromise));
    }

    describe('tag-strategy route', () => {
        let guard: TagStrategyGuardService;
        beforeEach(() => {
            guard = TestBed.inject(TagStrategyGuardService);
        });

        it('should allow users with tag retention permission only', async () => {
            permissions = of([true, false]);
            expect(await resolve(guard.canActivate(tagStrategy, state))).toBe(
                true
            );
        });

        it('should allow users with immutable tag permission only', async () => {
            permissions = of([false, true]);
            expect(await resolve(guard.canActivate(tagStrategy, state))).toBe(
                true
            );
        });

        it('should redirect users without any tag feature permission', async () => {
            permissions = of([false, false]);
            const result = await resolve(guard.canActivate(tagStrategy, state));
            expect(router.serializeUrl(result as UrlTree)).toEqual(
                '/harbor/projects'
            );
        });

        it('should redirect when permissions cannot be loaded', async () => {
            permissions = throwError(() => new Error('failed'));
            const result = await resolve(guard.canActivate(tagStrategy, state));
            expect(router.serializeUrl(result as UrlTree)).toEqual(
                '/harbor/projects'
            );
        });
    });

    describe('default tab', () => {
        let guard: TagStrategyDefaultTabGuardService;
        beforeEach(() => {
            guard = TestBed.inject(TagStrategyDefaultTabGuardService);
        });

        it('should open tag retention when it is permitted', async () => {
            permissions = of([true, true]);
            const result = await resolve(guard.canActivate(defaultTab, state));
            expect(router.serializeUrl(result as UrlTree)).toEqual(
                '/harbor/projects/1/tag-strategy/tag-retention'
            );
        });

        it('should open immutable tag without tag retention permission', async () => {
            permissions = of([false, true]);
            const result = await resolve(guard.canActivate(defaultTab, state));
            expect(router.serializeUrl(result as UrlTree)).toEqual(
                '/harbor/projects/1/tag-strategy/immutable-tag'
            );
        });

        it('should redirect users without any tag feature permission', async () => {
            permissions = of([false, false]);
            const result = await resolve(guard.canActivate(defaultTab, state));
            expect(router.serializeUrl(result as UrlTree)).toEqual(
                '/harbor/projects'
            );
        });
    });
});
