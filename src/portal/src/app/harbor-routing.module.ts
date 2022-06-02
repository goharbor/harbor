// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the 'License');
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an 'AS IS' BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { NgModule } from '@angular/core';
import {
    PreloadAllModules,
    RouteReuseStrategy,
    RouterModule,
    Routes,
} from '@angular/router';
import { AuthCheckGuard } from './shared/router-guard/auth-user-activate.service';
import { SignInGuard } from './shared/router-guard/sign-in-guard-activate.service';
import { OidcGuard } from './shared/router-guard/oidc-guard-active.service';
import { HarborRouteReuseStrategy } from './route-reuse-strategy/harbor-route-reuse-strategy';

const harborRoutes: Routes = [
    { path: '', redirectTo: 'harbor', pathMatch: 'full' },
    {
        path: 'account',
        loadChildren: () =>
            import('./account/account.module').then(m => m.AccountModule),
    },
    {
        path: 'devcenter-api-2.0',
        loadChildren: () =>
            import('./dev-center/dev-center.module').then(
                m => m.DeveloperCenterModule
            ),
    },
    {
        path: 'oidc-onboard',
        canActivate: [OidcGuard, SignInGuard],
        loadChildren: () =>
            import('./oidc-onboard/oidc-onboard.module').then(
                m => m.OidcOnboardModule
            ),
    },
    {
        path: 'license',
        loadChildren: () =>
            import('./license/license.module').then(m => m.LicenseModule),
    },
    {
        path: 'harbor',
        canActivateChild: [AuthCheckGuard],
        loadChildren: () =>
            import('./base/base.module').then(m => m.BaseModule),
    },
    {
        path: '**',
        loadChildren: () =>
            import('./not-found/not-found.module').then(m => m.NotFoundModule),
    },
];

@NgModule({
    providers: [
        { provide: RouteReuseStrategy, useClass: HarborRouteReuseStrategy },
    ],
    imports: [
        RouterModule.forRoot(harborRoutes, {
            onSameUrlNavigation: 'reload',
            preloadingStrategy: PreloadAllModules,
            relativeLinkResolution: 'legacy',
        }),
    ],
    exports: [RouterModule],
})
export class HarborRoutingModule {}
