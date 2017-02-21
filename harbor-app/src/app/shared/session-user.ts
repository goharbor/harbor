//Define the session user
export class SessionUser {
    user_id: number;
    username: string;
    email: string;
    realname: string;
    role_name?: string;
    role_id?: number;
    has_admin_role?: number;
    comment: string;
}