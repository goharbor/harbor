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
import { LabelFilterComponent } from './label-filter/label-filter.component';
import { LabelMarkerComponent } from './label-marker/label-marker.component';
import { ListChartVersionsComponent } from './list-chart-versions/list-chart-versions.component';
import { ChartVersionComponent } from './list-chart-versions/helm-chart-versions-detail/helm-chart-version.component';
import { ListChartsComponent } from './list-charts.component';
import { HelmChartComponent } from './list-charts-detail/helm-chart.component';
import { ChartDetailDependencyComponent } from './helm-chart-detail/chart-detail/chart-detail-dependency.component';
import { ChartDetailSummaryComponent } from './helm-chart-detail/chart-detail/chart-detail-summary.component';
import { ChartDetailValueComponent } from './helm-chart-detail/chart-detail/chart-detail-value.component';
import { ChartDetailComponent } from './helm-chart-detail/chart-detail/chart-detail.component';
import { HelmChartDetailComponent } from './helm-chart-detail/chart-detail.component';
import { SharedModule } from '../../shared/shared.module';
import { HelmChartDefaultService, HelmChartService } from './helm-chart.service';

@NgModule({
    imports: [SharedModule],
    declarations: [
        LabelFilterComponent,
        LabelMarkerComponent,
        ListChartVersionsComponent,
        ChartVersionComponent,
        ListChartsComponent,
        HelmChartComponent,
        ChartDetailDependencyComponent,
        ChartDetailSummaryComponent,
        ChartDetailValueComponent,
        ChartDetailComponent,
        HelmChartDetailComponent,
    ],
    providers: [
        { provide: HelmChartService, useClass: HelmChartDefaultService }],
    exports: [
        LabelFilterComponent,
        LabelMarkerComponent,
        ListChartVersionsComponent,
        ChartVersionComponent,
        ListChartsComponent,
        HelmChartComponent,
        ChartDetailDependencyComponent,
        ChartDetailSummaryComponent,
        ChartDetailValueComponent,
        ChartDetailComponent,
        HelmChartDetailComponent
    ]
})
export class HelmChartModule { }
