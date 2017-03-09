import { Injectable } from '@angular/core';
import { Http, URLSearchParams } from '@angular/http';

import { Repository } from './repository';
import { Observable } from 'rxjs/Observable'

@Injectable()
export class RepositoryService {
  
  constructor(private http: Http){}

  listRepositories(projectId: number, repoName: string, page?: number, pageSize?: number): Observable<any> {
    console.log('List repositories with project ID:' + projectId);
    let params = new URLSearchParams();
    params.set('page', page + '');
    params.set('page_size', pageSize + '');
    return this.http
               .get(`/api/repositories?project_id=${projectId}&q=${repoName}&detail=1`, {search: params})
               .map(response=>response)
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