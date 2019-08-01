export class Robot {
  project_id: number;
  id: number;
  name: string;
  description: string;
  expires_at: number;
  disabled: boolean;
  access: {
    isPullImage: boolean;
    isPushOrPullImage: boolean;
    isPushChart: boolean;
    isPullChart: boolean;
  };


  constructor () {
    this.access = <any>{};
    // this.access[0].action = true;
    this.access.isPullImage = false;
    this.access.isPushOrPullImage = true;
    this.access.isPushChart = false;
    this.access.isPullChart = false;
  }
}

