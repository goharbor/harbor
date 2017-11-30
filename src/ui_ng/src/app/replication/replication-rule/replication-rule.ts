import {Project} from "../../project/project";
/**
 * Created by pengf on 12/7/2017.
 */

export class Target {
    id: 0;
    endpoint: 'string';
    name: 'string';
    username: 'string';
    password: 'string';
    type: 0;
    insecure: true;
    creation_time: 'string';
    update_time: 'string';
}

export class Filter {
    kind: string;
    pattern: string;
    constructor(kind: string, pattern: string) {
        this.kind = kind;
        this.pattern = pattern;
    }
}

export class Trigger {
    kind: string;
    schedule_param: any | {
        [key: string]: any | any[];
    };
    constructor(kind: string, param: any | { [key: string]: any | any[]; }) {
        this.kind = kind;
        this.schedule_param = param;
    }
}

export interface ReplicationRule  {
    name: string;
    description: string;
    projects: Project[];
    targets: Target[] ;
    trigger: Trigger ;
    filters: Filter[] ;
    replicate_existing_image_now?: boolean;
    replicate_deletion?: boolean;
}

