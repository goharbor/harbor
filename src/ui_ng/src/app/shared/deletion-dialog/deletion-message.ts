export class DeletionMessage {
    public constructor(title: string, message: string, data: any){
        this.title = title;
        this.message = message;
        this.data = data;
    }
    title: string;
    message: string;
    data: any;
}