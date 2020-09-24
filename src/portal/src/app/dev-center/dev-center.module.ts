import { NgModule } from "@angular/core";
import { RouterModule } from "@angular/router";
import { CommonModule } from "@angular/common";
import { ClarityModule } from '@clr/angular';
import { SharedModule } from '../shared/shared.module';
import { DevCenterComponent } from "./dev-center.component";
import { DevCenterOtherComponent } from "./dev-center-other.component";


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
        DevCenterOtherComponent,
    ],
})
export class DeveloperCenterModule {}
