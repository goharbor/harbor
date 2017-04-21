import { Component, OnInit, Input } from '@angular/core';
import { SystemInfoService } from './providers/system-info.service';

@Component({
  selector: 'hbr-system',
  template: `
    <pre>
      {{info}}
    </pre>
  `,
  styles: [],
  providers: [SystemInfoService]
})

export class SystemComponent implements OnInit {
  _systemInfo: string = "Loading...";

  constructor(private systemService: SystemInfoService) { }

  public get info(): string {
    return this._systemInfo;
  }

  ngOnInit() {
    this.getInfo();
  }

  getInfo(): void {
    this.systemService.getSystemInfo()
    .then((res: any) => this._systemInfo = JSON.stringify(res))
    .catch(error => console.error("Retrieve system info error: ", error));
  }

}
