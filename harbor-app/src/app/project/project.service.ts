import { Injectable } from '@angular/core';

import { Http, Headers, RequestOptions } from '@angular/http';
import { Project } from './project';

import { BaseService } from '../service/base.service';

import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/map';

@Injectable()
export class ProjectService extends BaseService {
  
  headers = new Headers({'Content-type': 'application/json'});
  options = new RequestOptions({'headers': this.headers});

  constructor(private http: Http) {
    super();
  }

  listProjects(name: string, isPublic: number): Observable<Project[]>{    
    return this.http
               .get(`/ng/api/projects?project_name=${name}&is_public=${isPublic}`, this.options)
               .map(response=>response.json())
               .catch(this.handleError);
  }

  createProject(name: string, isPublic: number): Observable<any> {
    return this.http
               .post(`/ng/api/projects`,
                JSON.stringify({'project_name': name, 'public': (isPublic ? 1 : 0)})
                , this.options)
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }

}