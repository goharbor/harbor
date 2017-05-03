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
import { Http, URLSearchParams, Response } from '@angular/http';

import { Repository } from './repository';
import { Tag } from './tag';
import { VerifiedSignature } from './verified-signature';

import { Observable } from 'rxjs/Observable'
import 'rxjs/add/observable/of';
import 'rxjs/add/operator/mergeMap';

@Injectable()
export class RepositoryService {
  
  constructor(private http: Http){}

  listRepositories(projectId: number, repoName: string, page?: number, pageSize?: number): Observable<any> {
    let params = new URLSearchParams();
    if(page && pageSize) {
      params.set('page', page + '');
      params.set('page_size', pageSize + '');
    }
    return this.http
               .get(`/api/repositories?project_id=${projectId}&q=${repoName}&detail=1`, {search: params})
               .map(response=>response)
               .catch(error=>Observable.throw(error));
  }

  listTags(repoName: string): Observable<Tag[]> {
    return this.http
               .get(`/api/repositories/${repoName}/tags?detail=1`)
               .map(response=>response.json())
               .catch(error=>Observable.throw(error));
  }

  listNotarySignatures(repoName: string): Observable<VerifiedSignature[]> {
    return this.http
               .get(`/api/repositories/${repoName}/signatures`)
               .map(response=>response.json())
               .catch(error=>Observable.throw(error));
  }

  listTagsWithVerifiedSignatures(repoName: string): Observable<Tag[]> {
    return this.listTags(repoName)
               .map(res=>res)
               .flatMap(tags=>{
                 return this.listNotarySignatures(repoName).map(signatures=>{
                    tags.forEach(t=>{
                      for(let i = 0; i < signatures.length; i++) {
                        if(signatures[i].tag === t.tag) {
                          t.signed = 1;
                          break;
                        }
                      }
                    });
                    return tags;
                  })
                  .catch(error=>{
                    return Observable.of(tags);
                  })
               })
               .catch(error=>Observable.throw(error));
  }

  deleteRepository(repoName: string): Observable<any> {
    return this.http
               .delete(`/api/repositories/${repoName}/tags`)
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }

  deleteRepoByTag(repoName: string, tag: string): Observable<any> {
    return this.http
               .delete(`/api/repositories/${repoName}/tags/${tag}`)
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }

}