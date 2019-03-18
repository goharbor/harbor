
import {Injectable} from "@angular/core";
import {Http} from "@angular/http";
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError } from "rxjs";
@Injectable()
export class SkinableConfig {
    customSkinData: {[key: string]: any};
    constructor(private http: Http) {}

    public getCustomFile(): Observable<any> {
       return this.http.get('setting.json')
           .pipe(map(response => { this.customSkinData = response.json(); return this.customSkinData; })
           , catchError((error: any) => {
                console.error('custom skin json file load failed');
                return observableThrowError(error);
           }));
    }

    public getSkinConfig() {
        return this.customSkinData;
    }

    public getProject() {
        if (this.customSkinData) {
            return this.customSkinData.project;
        } else {
            return null;
        }
    }
}
