
import { CookieService } from "ngx-cookie";

export class XSRFStrategyToBeUsed {
    private cookieName: string = '_xsrf';
    private headerName: string = 'X-Xsrftoken';
    constructor(

        private cookieService: CookieService
    ) {
    }

    configureRequest(req: Request): void {
        console.log('Configure request');
        let token = null;
        const csrfCookie = this.cookieService.get(this.cookieName);
        if (csrfCookie) {
            token = atob(csrfCookie.split("|")[0]);
        }
        return req.headers.set(this.headerName, token);
    }
}
