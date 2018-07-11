import { Type } from '@angular/core';
import { HelmChartComponent } from './helm-chart.component';
import { ChartVersionComponent } from './versions/helm-chart-version.component';
import { ChartDetailComponent } from './chart-detail/chart-detail.component';
import { ChartDetailSummaryComponent } from './chart-detail/chart-detail-summary.component';
import { ChartDetailDependencyComponent } from './chart-detail/chart-detail-dependency.component';
import { ChartDetailValueComponent } from './chart-detail/chart-detail-value.component';

export * from "./helm-chart.component";
export * from "./versions/helm-chart-version.component";
export * from "./chart-detail/chart-detail.component";
export * from "./chart-detail/chart-detail-summary.component";
export * from "./chart-detail/chart-detail-dependency.component";
export * from "./chart-detail/chart-detail-value.component";

export const HELMCHART_DIRECTIVE: Type<any>[] = [
    HelmChartComponent,
    ChartVersionComponent,
    ChartDetailComponent,
    ChartDetailSummaryComponent,
    ChartDetailDependencyComponent,
    ChartDetailValueComponent,
];
