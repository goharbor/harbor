import { ScannerMetadata } from "./scanner-metadata";

export class Scanner {
    name?: string;
    description?: string;
    uuid?: string;
    url?: string;
    auth?: string;
    access_credential?: string;
    adapter?: string;
    disabled?: boolean;
    is_default?: boolean;
    skip_certVerify?: boolean;
    use_internal_addr?: boolean;
    create_time?: any;
    update_time?: any;
    vendor?: string;
    version?: string;
    metadata?: ScannerMetadata;
    loadingMetadata?: boolean;
    health?: string;
    constructor() {
    }
}
