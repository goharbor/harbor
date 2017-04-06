export class Volumes {
    constructor(){
        this.storage = new Storage();
    }

    storage: Storage;
}

export class Storage {
    constructor(){
        this.total = 0;
        this.free = 0;
    }

    total: number;
    free: number;
}