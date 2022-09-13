import { HttpParams } from '@angular/common/http';

/**
 * Wrap the class 'URLSearchParams' for future extending requirements.
 * Currently no extra methods provided.
 *
 **
 * class RequestQueryParams
 * extends {URLSearchParams}
 */
export class RequestQueryParams extends HttpParams {
    constructor() {
        super();
    }
}
