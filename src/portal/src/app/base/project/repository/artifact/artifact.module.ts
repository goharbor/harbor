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
import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { SharedModule } from '../../../../shared/shared.module';
import { ArtifactListPageComponent } from './artifact-list-page/artifact-list-page.component';
import { ArtifactListTabComponent } from './artifact-list-page/artifact-list/artifact-list-tab/artifact-list-tab.component';
import { ArtifactSummaryComponent } from './artifact-summary.component';
import { ArtifactLabelComponent } from './artifact-label/artifact-label.component';
import { ArtifactTagComponent } from './artifact-tag/artifact-tag.component';
import { ArtifactCommonPropertiesComponent } from './artifact-common-properties/artifact-common-properties.component';
import { ArtifactAdditionsComponent } from './artifact-additions/artifact-additions.component';
import { ValuesComponent } from './artifact-additions/values/values.component';
import { SummaryComponent } from './artifact-additions/summary/summary.component';
import { DependenciesComponent } from './artifact-additions/dependencies/dependencies.component';
import { BuildHistoryComponent } from './artifact-additions/build-history/build-history.component';
import { ArtifactVulnerabilitiesComponent } from './artifact-additions/artifact-vulnerabilities/artifact-vulnerabilities.component';
import { ArtifactFilesComponent } from './artifact-additions/files/files.component';
import { ArtifactLicenseComponent } from './artifact-additions/license/license.component';
import { ArtifactSbomComponent } from './artifact-additions/artifact-sbom/artifact-sbom.component';
import { ArtifactDefaultService, ArtifactService } from './artifact.service';
import { ArtifactDetailRoutingResolverService } from '../../../../services/routing-resolvers/artifact-detail-routing-resolver.service';
import { ResultBarChartComponent } from './vulnerability-scanning/result-bar-chart.component';
import { ResultSbomComponent } from './sbom-scanning/sbom-scan.component';
import { ResultTipHistogramComponent } from './vulnerability-scanning/result-tip-histogram/result-tip-histogram.component';
import { HistogramChartComponent } from './vulnerability-scanning/histogram-chart/histogram-chart.component';
import { ArtifactInfoComponent } from './artifact-list-page/artifact-list/artifact-info/artifact-info.component';
import { SubAccessoriesComponent } from './artifact-list-page/artifact-list/artifact-list-tab/sub-accessories/sub-accessories.component';
import { ArtifactListPageService } from './artifact-list-page/artifact-list-page.service';
import { CopyArtifactComponent } from './artifact-list-page/artifact-list/artifact-list-tab/copy-artifact/copy-artifact.component';
import { CopyDigestComponent } from './artifact-list-page/artifact-list/artifact-list-tab/copy-digest/copy-digest.component';
import { ArtifactFilterComponent } from './artifact-list-page/artifact-list/artifact-list-tab/artifact-filter/artifact-filter.component';
import { PullCommandComponent } from './artifact-list-page/artifact-list/artifact-list-tab/pull-command/pull-command.component';
import { SbomTipHistogramComponent } from './sbom-scanning/sbom-tip-histogram/sbom-tip-histogram.component';

const routes: Routes = [
    {
        path: ':repo',
        component: ArtifactListPageComponent,
        children: [
            {
                path: 'info-tab',
                component: ArtifactInfoComponent,
            },
            {
                path: 'artifacts-tab',
                component: ArtifactListTabComponent,
            },
            { path: '', redirectTo: 'artifacts-tab', pathMatch: 'full' },
        ],
    },
    {
        path: ':repo',
        component: ArtifactListPageComponent,
        children: [
            {
                path: 'artifacts-tab/depth/:depth',
                component: ArtifactListTabComponent,
            },
        ],
    },
    {
        path: ':repo/artifacts-tab/artifacts/:digest',
        component: ArtifactSummaryComponent,
        resolve: {
            artifactResolver: ArtifactDetailRoutingResolverService,
        },
    },
    {
        path: ':repo/artifacts-tab/depth/:depth/artifacts/:digest',
        component: ArtifactSummaryComponent,
        resolve: {
            artifactResolver: ArtifactDetailRoutingResolverService,
        },
    },
];
@NgModule({
    declarations: [
        ArtifactListPageComponent,
        ArtifactListTabComponent,
        ArtifactSummaryComponent,
        ArtifactLabelComponent,
        ArtifactLicenseComponent,
        ArtifactFilesComponent,
        ArtifactTagComponent,
        ArtifactCommonPropertiesComponent,
        ArtifactAdditionsComponent,
        ValuesComponent,
        SummaryComponent,
        DependenciesComponent,
        BuildHistoryComponent,
        ArtifactSbomComponent,
        ArtifactVulnerabilitiesComponent,
        ResultBarChartComponent,
        ResultSbomComponent,
        SbomTipHistogramComponent,
        ResultTipHistogramComponent,
        HistogramChartComponent,
        ArtifactInfoComponent,
        SubAccessoriesComponent,
        CopyArtifactComponent,
        CopyDigestComponent,
        ArtifactFilterComponent,
        PullCommandComponent,
    ],
    imports: [RouterModule.forChild(routes), SharedModule],
    providers: [
        ArtifactListPageService,
        { provide: ArtifactService, useClass: ArtifactDefaultService },
    ],
})
export class ArtifactModule {}
