export class UserGroup {
  id?: number;
  group_name?: string;
  group_type: number;
  ldap_group_dn?: string;

  constructor() {
    {
      this.group_type = 1;
    }
  }
}
