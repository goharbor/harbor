import { NgModule } from "@angular/core";
import { RouterModule } from "@angular/router";
import { CommonModule } from "@angular/common";
import { ClarityModule } from '@clr/angular';
import { SharedModule } from '../shared/shared.module';
import { DevCenterComponent } from "./dev-center.component";


@NgModule({
    imports: [
        CommonModule,
        SharedModule,
        RouterModule.forChild([{
            path: "**",
            component: DevCenterComponent,
        }]),
        ClarityModule,
    ],
    declarations: [
        DevCenterComponent,
    ],
})
export class DeveloperCenterModule {}
