/*
 {
    "log_id": 3,
    "user_id": 0,
    "project_id": 0,
    "repo_name": "library/mysql",
    "repo_tag": "5.6",
    "guid": "",
    "operation": "push",
    "op_time": "2017-02-14T09:22:58Z",
    "username": "admin",
    "keywords": "",
    "BeginTime": "0001-01-01T00:00:00Z",
    "begin_timestamp": 0,
    "EndTime": "0001-01-01T00:00:00Z",
    "end_timestamp": 0
  }
*/
export class AuditLog {
  log_id: number;
  project_id: number;
  username: string;
  repo_name: string;
  repo_tag: string;
  operation: string;
  op_time: Date;
  begin_timestamp: number = 0;
  end_timestamp: number = 0;
  keywords: string;
  page: number;
  page_size: number;
}