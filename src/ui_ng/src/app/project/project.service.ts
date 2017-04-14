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
import { Injectable } from '@angular/core';

import { Http, Headers, RequestOptions, Response, URLSearchParams } from '@angular/http';
import { Project } from './project';

import { Message } from '../global-message/message';

import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/map';
import 'rxjs/add/observable/throw';



@Injectable()
export class ProjectService {
  
  headers = new Headers({'Content-type': 'application/json'});
  options = new RequestOptions({'headers': this.headers});

  constructor(private http: Http) {}

  getProject(projectId: number): Observable<any> {
    return this.http
               .get(`/api/projects/${projectId}`)
               .map(response=>response.json())
               .catch(error=>Observable.throw(error));
  }

  listProjects(name: string, isPublic: number, page?: number, pageSize?: number): Observable<any>{    
    let params = new URLSearchParams();
    params.set('page', page + '');
    params.set('page_size', pageSize + '');
    return this.http
               .get(`/api/projects?project_name=${name}&is_public=${isPublic}`, {search: params})
               .map(response=>response)
               .catch(error=>Observable.throw(error));
  }

  createProject(name: string, isPublic: number): Observable<any> {
    return this.http
               .post(`/api/projects`,
                JSON.stringify({'project_name': name, 'public': isPublic})
                , this.options)
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }

  toggleProjectPublic(projectId: number, isPublic: number): Observable<any> {
    return this.http 
               .put(`/api/projects/${projectId}/publicity`, { 'public': isPublic }, this.options)
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }

  deleteProject(projectId: number): Observable<any> {
    return this.http
               .delete(`/api/projects/${projectId}`)
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }

  checkProjectExists(projectName: string): Observable<any> {
    return this.http
               .head(`/api/projects/?project_name=${projectName}`)
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }
   
  checkProjectMember(projectId: number): Observable<any> {
    return this.http
               .get(`/api/projects/${projectId}/members`)
               .map(response=>response.json())
               .catch(error=>Observable.throw(error));
  }
  
}