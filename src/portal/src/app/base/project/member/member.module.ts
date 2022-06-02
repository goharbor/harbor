import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { SharedModule } from '../../../shared/shared.module';
import { MemberComponent } from './member.component';
import { AddMemberComponent } from './add-member/add-member.component';
import { AddGroupComponent } from './add-group/add-group.component';

const routes: Routes = [
    {
        path: '',
        component: MemberComponent,
    },
];
@NgModule({
    declarations: [MemberComponent, AddMemberComponent, AddGroupComponent],
    imports: [RouterModule.forChild(routes), SharedModule],
})
export class MemberModule {}
