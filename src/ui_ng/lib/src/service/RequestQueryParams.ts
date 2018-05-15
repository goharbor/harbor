import { URLSearchParams } from "@angular/http";

/**
 * Wrap the class 'URLSearchParams' for future extending requirements.
 * Currently no extra methods provided.
 *
 * @export
 * @class RequestQueryParams
 * @extends {URLSearchParams}
 */
export class RequestQueryParams extends URLSearchParams {
  constructor() {
    super();
  }
}
