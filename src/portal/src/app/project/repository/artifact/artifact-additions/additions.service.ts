import { Injectable } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { Observable } from "rxjs";

@Injectable({
  providedIn: 'root',
})
export class AdditionsService {
  constructor(private http: HttpClient) {
  }

  getDetailByLink(link: string, shouldReturnText?: boolean): Observable<any> {
    if (shouldReturnText) {
      return this.http.get(link, { observe: 'body', responseType: 'text'} );
    }
    return  this.http.get(link);
  }
}
