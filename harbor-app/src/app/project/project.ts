/*
  [
    {
        "project_id": 1,
        "owner_id": 1,
        "name": "library",
        "creation_time": "2017-02-10T07:57:56Z",
        "creation_time_str": "",
        "deleted": 0,
        "owner_name": "",
        "public": 1,
        "Togglable": true,
        "update_time": "2017-02-10T07:57:56Z",
        "current_user_role_id": 1,
        "repo_count": 0
    }
  ]
*/
export class Project { 
    project_id: number;
    owner_id: number;
    name: string;
    creation_time: Date;
    creation_time_str: string;
    deleted: number;
    owner_name: string;
    public: number;
    Togglable: boolean;
    update_time: Date;
    current_user_role_id: number;
    repo_count: number;
}