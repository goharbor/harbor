import { Project } from '../../project/project';
import { Repository } from '../../repository/repository';

export class SearchResults {
    constructor(){
        this.project = [];
        this.repository = [];
    }

    project: Project[];
    repository: Repository[];
}