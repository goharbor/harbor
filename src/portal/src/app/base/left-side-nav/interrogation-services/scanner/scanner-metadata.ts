export class ScannerMetadata {
    scanner?: {
        name?: string;
        vendor?: string;
        version?: string;
    };
    capabilities?: [{
        consumes_mime_types?: Array<string>;
        produces_mime_types?: Array<string>;
    }];
    properties?: {
      [key: string]: string;
    };
    constructor() {
    }
}
