/*
 {
    "id": 1,
    "endpoint": "http://10.117.4.151",
    "name": "target_01",
    "username": "admin",
    "password": "Harbor12345",
    "type": 0,
    "creation_time": "2017-02-24T06:41:52Z",
    "update_time": "2017-02-24T06:41:52Z"
  }
*/

export class Target {
  id: number;
  endpoint: string;
  name: string;
  username: string;
  password: string;
  type: number;
  creation_time: Date;
  update_time: Date;
}