import { Injectable } from '@angular/core';
import { ActivatedRouteSnapshot, RouterStateSnapshot } from '@angular/router';
import { Observable } from 'rxjs';
import { UserPrivilegeServeItem } from 'src/app/shared/services/interface';
import { MemberPermissionGuard } from '../../../shared/router-guard/member-permission-guard-activate.service';

@Injectable({
    providedIn: 'root',
})
export class TagFeatureGuardService {
    constructor(private memberPermissionGuard: MemberPermissionGuard) {}

    canActivate(
        route: ActivatedRouteSnapshot,
        state: RouterStateSnapshot
    ): Observable<boolean> | boolean {
        const projectId = route.parent.parent.parent.params['id'];
        const permission = route.data.permissionParam as UserPrivilegeServeItem;
        return this.memberPermissionGuard.checkPermission(
            projectId,
            permission
        );
    }
}
