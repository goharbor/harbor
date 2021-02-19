import { NgModule } from '@angular/core';
import { RouterModule, Routes } from "@angular/router";
import { SharedModule } from "../../../shared/shared.module";
import { MemberComponent } from "./member.component";
import { AddMemberComponent } from "./add-member/add-member.component";
import { AddHttpAuthGroupComponent } from "./add-http-auth-group/add-http-auth-group.component";
import { AddGroupComponent } from "./add-group/add-group.component";
import { MemberService } from "./member.service";

const routes: Routes = [
  {
    path: '',
    component: MemberComponent
  }
];
@NgModule({
  declarations: [
    MemberComponent,
    AddMemberComponent,
    AddHttpAuthGroupComponent,
    AddGroupComponent
  ],
  imports: [
    RouterModule.forChild(routes),
    SharedModule
  ],
  providers: [
    MemberService
  ]
})
export class MemberModule { }
