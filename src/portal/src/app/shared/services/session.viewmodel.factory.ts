import { Injectable } from '@angular/core';
import { SessionUser, SessionUserBackend } from '../entities/session-user';
import { clone } from '../units/utils';

@Injectable({
    providedIn: 'root',
})
export class SessionViewmodelFactory {
    constructor() {}
    // view model need
    getCurrentUser(currentUser: SessionUserBackend): SessionUser {
        return {
            user_id: currentUser.user_id,
            username: currentUser.username,
            email: currentUser.email,
            realname: currentUser.realname,
            role_name: currentUser.role_name,
            role_id: currentUser.role_id,
            comment: currentUser.comment,
            oidc_user_meta: currentUser.oidc_user_meta,
            has_admin_role:
                currentUser.admin_role_in_auth || currentUser.sysadmin_flag,
        };
    }
}
