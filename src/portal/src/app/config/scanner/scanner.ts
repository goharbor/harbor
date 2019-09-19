export class Scanner {
    name?: string;
    description?: string;
    uid?: string;
    url?: string;
    auth?: string;
    accessCredential?: string;
    adapter?: string;
    disabled?: boolean;
    isDefault?: boolean;
    skipCertVerify?: boolean;
    createTime?: any;
    updateTime?: any;
    vendor?: string;
    version?: string;
    constructor() {
        this.adapter = "Clair";
        this.vendor = "Harbor";
        this.version = "1.0.0";
    }
}
