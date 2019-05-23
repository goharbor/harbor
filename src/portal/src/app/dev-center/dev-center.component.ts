import { AfterViewInit, Component, ElementRef, OnInit } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { throwError as observableThrowError, Observable } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
import { Title } from '@angular/platform-browser';
import { TranslateService } from '@ngx-translate/core';


const SwaggerUI = require('swagger-ui');
@Component({
  selector: 'dev-center',
  templateUrl: 'dev-center.component.html',
  viewProviders: [Title],
  styleUrls: ['dev-center.component.scss']
})
export class DevCenterComponent implements AfterViewInit, OnInit {
  private ui: any;
  private host: any;
  private json: any;
  constructor(
    private el: ElementRef,
    private http: HttpClient,
    private translate: TranslateService,
    private titleService: Title) {
  }

  ngOnInit() {
    this.setTitle("APP_TITLE.HARBOR_SWAGGER");
  }


  public setTitle( key: string) {
    this.translate.get(key).subscribe((res: string) => {
      this.titleService.setTitle(res);
  });
  }

  ngAfterViewInit() {
    this.http.get("/swagger.json")
    .pipe(catchError(error => observableThrowError(error)))
    .subscribe(json => {
      json['host'] = window.location.host;
      const protocal = window.location.protocol;
      json['schemes'] = [protocal.replace(":", "")];
      let ui = SwaggerUI({
        spec: json,
        domNode: this.el.nativeElement.querySelector('.swagger-container'),
        deepLinking: true,
        presets: [
          SwaggerUI.presets.apis
        ],
      });
    });
  }
}
