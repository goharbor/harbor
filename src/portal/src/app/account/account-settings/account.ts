export class ResetSecret {
    input_secret: string;
    confirm_secret: string;
    constructor() {
        this.confirm_secret = "";
        this.input_secret = "";
    }
}
