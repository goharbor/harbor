/*
{
  "user_id": 1,
  "username": "admin",
  "email": "",
  "password": "",
  "realname": "",
  "comment": "",
  "deleted": 0,
  "role_name": "projectAdmin",
  "role_id": 1,
  "has_admin_role": 0,
  "reset_uuid": "",
  "creation_time": "0001-01-01T00:00:00Z",
  "update_time": "0001-01-01T00:00:00Z"
}
*/

export class Member {
  user_id: number;
  username: string;
  role_name: string;
  has_admin_role: number;
  role_id: number;
}