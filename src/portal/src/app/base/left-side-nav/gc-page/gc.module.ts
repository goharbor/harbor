import { NgModule } from "@angular/core";
import { RouterModule, Routes } from "@angular/router";
import { GcPageComponent } from "./gc-page.component";
import { GcComponent } from "./gc/gc.component";
import { GcHistoryComponent } from "./gc/gc-history/gc-history.component";
import { GcRepoService } from "./gc/gc.service";
import { SharedModule } from "../../../shared/shared.module";
import { GcApiDefaultRepository, GcApiRepository } from "./gc/gc.api.repository";
import { GcViewModelFactory } from "./gc/gc.viewmodel.factory";

const routes: Routes = [
    {
        path: '',
        component: GcPageComponent
    }
];
@NgModule({
    imports: [
        SharedModule,
        RouterModule.forChild(routes)
    ],
    declarations: [
        GcPageComponent,
        GcComponent,
        GcHistoryComponent
    ],
    providers: [
        GcRepoService,
        {provide: GcApiRepository, useClass: GcApiDefaultRepository },
        GcViewModelFactory
    ]
})
export class GcModule {}
