/*
  {
    "id": 1,
    "status": "running",
    "repository": "library/mysql",
    "policy_id": 1,
    "operation": "transfer",
    "tags": null,
    "creation_time": "2017-02-24T06:44:04Z",
    "update_time": "2017-02-24T06:44:04Z"
  }

*/
export class Job {
  id: number;
  status: string;
  repository: string;
  policy_id: number;
  operation: string;
  tags: string;
  creation_time: Date;
  update_time: Date;
}