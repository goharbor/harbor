
import {Injectable} from "@angular/core";
import {Http} from "@angular/http";

@Injectable()
export class SkinableConfig {
    customSkinData: {[key: string]: any};
    constructor(private http: Http) {}

    public getCustomFile(): Promise<any> {
       return this.http.get('../static/setting.json')
           .toPromise()
           .then(response => { this.customSkinData = response.json(); return this.customSkinData; })
           .catch(error => {
               console.error('custom skin json file load failed');
           });
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
