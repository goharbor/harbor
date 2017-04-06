/*
  {
    "id": 1,
    "project_id": 1,
    "project_name": "library",
    "target_id": 1,
    "target_name": "target_01",
    "name": "sync_01",
    "enabled": 0,
    "description": "sync_01 desc.",
    "cron_str": "",
    "start_time": "0001-01-01T00:00:00Z",
    "creation_time": "2017-02-24T06:41:52Z",
    "update_time": "2017-02-24T06:41:52Z",
    "error_job_count": 0,
    "deleted": 0
  }
*/

export class Policy {
  id: number;
  project_id: number;
  project_name: string;
  target_id: number;
  target_name: string;
  name: string;
  enabled: number;
  description: string;
  cron_str: string;
  start_time: Date;
  creation_time: Date;
  update_time: Date;
  error_job_count: number;
  deleted: number;
}