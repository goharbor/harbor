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