import { throwError as observableThrowError, Observable } from "rxjs";

import { map, catchError } from "rxjs/operators";
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { Injectable } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { RobotApiRepository } from "./robot.api.repository";
import { Robot } from "./robot";
@Injectable()
export class RobotService {
  constructor(
    private http: HttpClient,
    private robotApiRepository: RobotApiRepository
  ) { }

  /** addRobotAccount
   * projectId
   * robot: Robot
   * projectName
   */
  public addRobotAccount(projectId: number, robot: Robot, projectName: string): Observable<any> {
    let access = [];
    if (robot.access.isPullImage) {
      access.push({ "resource": `/project/${projectId}/repository`, "action": "pull" });
    }
    if (robot.access.isPushOrPullImage) {
      access.push({ "resource": `/project/${projectId}/repository`, "action": "push" });
    }
    if (robot.access.isPullChart) {
      access.push({ "resource": `/project/${projectId}/helm-chart`, "action": "read" });
    }
    if (robot.access.isPushChart) {
      access.push({ "resource": `/project/${projectId}/helm-chart-version`, "action": "create" });
    }

    let param = {
      name: robot.name,
      description: robot.description,
      access
    };

    return this.robotApiRepository.postRobot(projectId, param);
  }

  public deleteRobotAccount(projectId, id): Observable<any> {
    return this.robotApiRepository.deleteRobot(projectId, id);
  }

  public listRobotAccount(projectId): Observable<any> {
    return this.robotApiRepository.listRobot(projectId);
  }

  public getRobotAccount(projectId, id): Observable<any> {
    return this.robotApiRepository.getRobot(projectId, id);
  }

  public toggleDisabledAccount(projectId, id, isDisabled): Observable<any> {
    let data = {
      Disabled: isDisabled
    };
    return this.robotApiRepository.toggleDisabledAccount(projectId, id, data);
  }
}
