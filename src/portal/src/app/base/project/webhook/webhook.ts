export class Webhook {
  id: number;
  name: string;
  project_id: number;
  description: string;
  targets: Target[];
  event_types: string[];
  creator: string;
  creation_time: Date;
  update_time: Date;
  enabled: boolean;
  constructor () {
    this.targets = [];
    this.targets.push(new Target());
    this.event_types = [];
    this.enabled = true;
  }
}

export class Target {
  type: string;
  address: string;
  attachment: string;
  auth_header: string;
  skip_cert_verify: boolean;

  constructor () {
    this.type = 'http';
    this.address = '';
    this.skip_cert_verify = true;
  }
}

export class LastTrigger {
  policy_name: string;
  enabled: boolean;
  event_type: string;
  creation_time: Date;
  last_trigger_time: Date;
}
