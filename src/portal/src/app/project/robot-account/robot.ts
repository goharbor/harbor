export class Robot {
  project_id: number;
  id: number;
  name: string;
  description: string;
  expires_at: number;
  disabled: boolean;
  access: {
    isPull: boolean;
    isPush: boolean;
  };


  constructor () {
    this.access = <any>{};
    // this.access[0].action = true;
    this.access.isPull = true;
    this.access.isPush = true;
  }
}

