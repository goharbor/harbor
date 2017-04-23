import { Observable } from 'rxjs/Observable';
import { RequestQueryParams } from './RequestQueryParams';
import { Tag } from './interface';
import { Injectable } from "@angular/core";
import 'rxjs/add/observable/of';

/**
 * Define the service methods to handle the repository tag related things.
 * 
 * @export
 * @abstract
 * @class TagService
 */
export abstract class TagService {
    /**
     * Get all the tags under the specified repository.
     * NOTES: If the Notary is enabled, the signatures should be included in the returned data.
     * 
     * @abstract
     * @param {string} repositoryName
     * @param {RequestQueryParams} [queryParams]
     * @returns {(Observable<Tag[]> | Tag[])}
     * 
     * @memberOf TagService
     */
    abstract getTags(repositoryName: string, queryParams?: RequestQueryParams): Observable<Tag[]> | Tag[];

    /**
     * Delete the specified tag.
     * 
     * @abstract
     * @param {string} repositoryName
     * @param {string} tag
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf TagService
     */
    abstract deleteTag(repositoryName: string, tag: string): Observable<any> | any;
}

/**
 * Implement default service for tag.
 * 
 * @export
 * @class TagDefaultService
 * @extends {TagService}
 */
@Injectable()
export class TagDefaultService extends TagService {
    public getTags(repositoryName: string, queryParams?: RequestQueryParams): Observable<Tag[]> | Tag[] {
        return Observable.of([]);
    }

    public deleteTag(repositoryName: string, tag: string): Observable<any> | any {
        return Observable.of({});
    }
}