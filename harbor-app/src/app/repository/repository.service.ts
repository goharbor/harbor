import { Injectable } from '@angular/core';
import { Http } from '@angular/http';

import { Repository } from './repository';
import { Observable } from 'rxjs/Observable'

@Injectable()
export class RepositoryService {
  
  constructor(private http: Http){}

  listRepositories(projectId: number, repoName: string): Observable<Repository[]> {
    console.log('List repositories with project ID:' + projectId);
    return this.http
               .get(`/api/repositories?project_id=${projectId}&q=${repoName}&detail=1`)
               .map(response=>response.json() as Repository[])
               .catch(error=>Observable.throw(error));
  }

  deleteRepository(repoName: string): Observable<any> {
    console.log('Delete repository with repo name:' + repoName);
    return this.http
               .delete(`/api/repositories?repo_name=${repoName}`)
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }
}