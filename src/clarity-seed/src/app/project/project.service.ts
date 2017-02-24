import { Injectable } from '@angular/core';

import { Http, Headers, RequestOptions } from '@angular/http';
import { Project } from './project';

import { BaseService } from '../service/base.service';

import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/map';
import 'rxjs/add/observable/throw';

const url_prefix = '';

@Injectable()
export class ProjectService extends BaseService {
  
  headers = new Headers({'Content-type': 'application/json'});
  options = new RequestOptions({'headers': this.headers});

  constructor(private http: Http) {
    super();
  }

  getProject(projectId: number): Promise<Project> {
    return this.http
               .get(url_prefix + `/api/projects/${projectId}`)
               .toPromise()
               .then(response=>response.json() as Project)
               .catch(error=>Observable.throw(error));
  }

  listProjects(name: string, isPublic: number): Observable<Project[]>{    
    return this.http
               .get(url_prefix + `/api/projects?project_name=${name}&is_public=${isPublic}`, this.options)
               .map(response=>response.json())
               .catch(this.handleError);
  }

  createProject(name: string, isPublic: number): Observable<any> {
    return this.http
               .post(url_prefix + `/api/projects`,
                JSON.stringify({'project_name': name, 'public': isPublic})
                , this.options)
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }

  toggleProjectPublic(projectId: number, isPublic: number): Observable<any> {
    return this.http 
               .put(url_prefix + `/api/projects/${projectId}/publicity`, { 'public': isPublic }, this.options)
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }

  deleteProject(projectId: number): Observable<any> {
    return this.http
               .delete(url_prefix + `/api/projects/${projectId}`)
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }
}