import {Project} from "../../project/project";
/**
 * Created by pengf on 12/7/2017.
 */

export class Target {
    id: number;
    endpoint: string;
    name: string;
    username: string;
    password: string;
    type: number;
    insecure: true;
    creation_time: string;
    update_time: string;
    constructor() {
        this.id = -1;
        this.endpoint = "";
        this.name = "";
        this.username = "";
        this.password = "";
        this.type = 0;
        this.insecure = true;
        this.creation_time = "";
        this.update_time = "";
    }
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
    id?: number;
    name: string;
    description: string;
    projects: Project[];
    targets: Target[] ;
    trigger: Trigger ;
    filters: Filter[] ;
    replicate_existing_image_now?: boolean;
    replicate_deletion?: boolean;
}

