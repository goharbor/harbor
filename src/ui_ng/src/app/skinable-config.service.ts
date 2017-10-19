
import {Injectable} from "@angular/core";
import {Http} from "@angular/http";
import {Observable} from "rxjs/Observable";
/**
 * Created by pengf on 9/15/2017.
 */

@Injectable()
export class SkinableConfig {
    customSkinData: {[key: string]: any};
    constructor(private http: Http) {}

    public getCustomFile(): Promise<any> {
       return this.http.get('../setting.json')
           .toPromise()
           .then(response => { this.customSkinData = response.json(); return this.customSkinData; })
           .catch(error => {
               console.error('custom skin json file load failed');
           });
    }

    public getSkinConfig() {
        return this.customSkinData;
    }

    public getProjects() {
        if (this.customSkinData) {
            return this.customSkinData.projects;
        }else {
            return null;
        }
    }
}