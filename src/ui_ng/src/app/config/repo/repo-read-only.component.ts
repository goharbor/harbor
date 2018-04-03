
import {Component, Input, ViewChild} from "@angular/core";
import {Configuration} from "harbor-ui";
import {NgForm} from "@angular/forms";

@Component({
    selector: 'repo-read-only',
    templateUrl: 'repo-read-only.html',
})
export class RepoReadOnlyComponent {

    @Input('repoConfig') currentConfig: Configuration = new Configuration();

    @ViewChild('repoConfigFrom') repoForm: NgForm;

    constructor() { }

    disabled(prop: any) {
        return !(prop && prop.editable);
    }

    setInsecureReadOnlyValue($event: any) {
        this.currentConfig.read_only.value = $event;
    }
}