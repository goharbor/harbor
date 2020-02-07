import { Label, Tag } from "../../services";

export class Artifact {
    id: number;
    type: string;
    repository: string;
    tags: Tag[];
    media_type: string;
    digest: string;
    size: number;
    upload_time?: string;
    // labels: string[];
    extra_attrs?: Map<string, string>;
    addition_links?: Map<string, string>;
    references: Reference[];
    scan_overview: any;
    labels: Label[];
    push_time: string;
    pull_time: string;
    isOpen?: boolean; // front
    referenceIndexOpenState?: boolean; // front
    referenceDigestOpenState?: boolean; // front
    hasReferenceArtifactList?: Artifact[] = []; // front
    noReferenceArtifactList?: Artifact[] = []; // front
    constructor(digestName, hasReference?) {
        this.id = 1;
        this.type = 'type';
        this.size = 1111111111;
        this.upload_time = '2020-01-06T09:40:08.036866579Z';
        this.digest = digestName;
        this.tags = [
            {
                id: '1',
                artifact_id: 1,
                name: 'tag1',
                upload_time: '2020-01-06T09:40:08.036866579Z'
            },
            {
                id: '2',
                artifact_id: 2,
                name: 'tag2',
                upload_time: '2020-01-06T09:40:08.036866579Z',
            },
        ];
        // tslint:disable-next-line: no-use-before-declare
        // this.references = [];
        this.references = hasReference ? [new Reference(1), new Reference(2)] : [];
    }
}
export class Reference {
    child_id: number;
    child_digest: string;
    parent_id: number;
    platform?: any; // json
    constructor(artifact_id) {
        this.child_id = artifact_id;
    }
}
