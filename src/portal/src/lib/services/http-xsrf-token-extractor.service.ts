import { Injectable } from "@angular/core";
import { HttpXsrfTokenExtractor } from "@angular/common/http";
import { CookieService } from "ngx-cookie";
@Injectable()
export class HttpXsrfTokenExtractorToBeUsed extends HttpXsrfTokenExtractor {
    constructor(
        private cookieService: CookieService,
    ) {
        super();
    }
    public getToken(): string | null {
        const csrfCookie = this.cookieService.get("_xsrf");
        if (csrfCookie) {
            return atob(csrfCookie.split("|")[0]);
        }
        return null;
    }
}
