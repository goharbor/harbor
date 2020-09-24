export class Robot {
  project_id: number;
  id: number;
  name: string;
  description: string;
  expires_at: number;
  disabled: boolean;
  creation_time?: Date;
  access: {
    isPullImage: boolean;
    isPushOrPullImage: boolean;
    isPushChart: boolean;
    isPullChart: boolean;
  };


  constructor () {
    this.access = <any>{};
    this.access.isPullImage = false;
    this.access.isPushOrPullImage = true;
    this.access.isPushChart = true;
    this.access.isPullChart = true;
  }
}

