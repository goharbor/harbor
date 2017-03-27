export class AppConfig {
    constructor(){
        //Set default value
        this.with_notary = false;
        this.with_admiral = false;
        this.admiral_endpoint = "";
        this.auth_mode = "db_auth";
        this.registry_url = "";
        this.project_creation_restriction = "everyone";
        this.self_registration = true;
        this.has_ca_root = false;
    }
    
    with_notary: boolean;
    with_admiral: boolean;
    admiral_endpoint: string;
    auth_mode: string;
    registry_url: string;
    project_creation_restriction: string;
    self_registration: boolean;
    has_ca_root: boolean;
}