export class DistributionProvider {
  name: string;
  icon: string;
  version: string;
  source: string;
  maintainers: string[];
  auth_mode: string;
}

export class DistributionInstance {
  id?: string;
  name?: string;
  endpoint?: string;
  description?: string;
  status?: string;
  enabled?: boolean;
  setup_timestamp?: Number;
  provider?: DistributionProvider | string;
  auth_mode?: string;
  auth_data?: AuthModeBasic | AuthModeOAuth;
}

export class DistributionHistory {
  image: string;
  start_time: string;
  finish_time: string;
  status: string;
  provider: string;
  instance: string;
}

export class AuthMode {
  static NONE = 'NONE';
  static BASIC = 'BASIC';
  static OAUTH = 'OAUTH';
  static CUSTOM = 'CUSTOM';
}

export class AuthModeBasic {
  username: string;
  password: string;
}
export class AuthModeOAuth {
  token: string;
}

export class QueryParam {
  page: number;
  pageSize: number;
  query: string;
}
