
import { CookieService } from "ngx-cookie";

export class XSRFStrategyToBeUsed {
    public cookieName: string = '_xsrf';
    public headerName: string = 'X-Xsrftoken';
    constructor(

        private cookieService: CookieService
    ) {
    }

    configureRequest(req: any): void {
        let token = null;
        const csrfCookie = this.cookieService.get(this.cookieName);
        if (csrfCookie) {
            token = atob(csrfCookie.split("|")[0]);
        }
        req.headers.set(this.headerName, token);
    }
}
