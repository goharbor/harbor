import { Component, OnInit } from '@angular/core';
import { SessionService } from "../shared/session.service";
import { ConfigurationService } from "../config/config.service";
import { MessageHandlerService } from "../shared/message-handler/message-handler.service";
import { finalize } from "rxjs/operators";
import { Configuration } from "../../lib/components/config/config";

@Component({
  selector: 'app-project-quotas',
  templateUrl: './project-quotas.component.html',
  styleUrls: ['./project-quotas.component.scss']
})
export class ProjectQuotasComponent implements OnInit {
  allConfig: Configuration = new Configuration();
  loading: boolean = false;
  constructor(private session: SessionService,
              private configService: ConfigurationService,
              private msgHandler: MessageHandlerService) { }

  ngOnInit() {
    let currentUser = this.session.getCurrentUser();
    if (currentUser && currentUser.has_admin_role) {
      this.retrieveConfig();
    }
  }

  refreshAllconfig() {
    this.retrieveConfig();
  }
  retrieveConfig(): void {
    this.loading = true;
    this.configService.getConfiguration()
        .pipe(finalize(() => this.loading = false))
        .subscribe((configurations: Configuration) => {
          this.allConfig = configurations;
        }, error => {
          this.msgHandler.handleError(error);
        });
  }
}
