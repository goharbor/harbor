import { Http, Response,} from '@angular/http';

export class BaseService {

  protected handleError(error: Response | any): Promise<any> {
     // In a real world app, we might use a remote logging infrastructure
    let errMsg: string;    
    console.log(typeof error);
    if (error instanceof Response) {
      const body = error.json() || '';
      const err = body.error || JSON.stringify(body);
      errMsg = `${error.status} - ${error.statusText || ''} ${err}`;
    } else {
      errMsg = error.message ? error.message : error.toString();
    }
    return Promise.reject(error);
  }
}