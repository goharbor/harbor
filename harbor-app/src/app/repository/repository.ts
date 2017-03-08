/*
  {
    "id": "2",
    "name": "library/mysql",
    "owner_id": 1,
    "project_id": 1,
    "description": "",
    "pull_count": 0,
    "star_count": 0,
    "tags_count": 1,
    "creation_time": "2017-02-14T09:22:58Z",
    "update_time": "0001-01-01T00:00:00Z"
  }
*/

export class Repository {
  id: number;
  name: string;
  owner_id: number;
  project_id: number;
  description: string;
  pull_count: number;
  start_count: number;
  tags_count: number;
  creation_time: Date;
  update_time: Date;

  constructor(name: string, tags_count: number) {
    this.name = name;
    this.tags_count = tags_count;
  }
}