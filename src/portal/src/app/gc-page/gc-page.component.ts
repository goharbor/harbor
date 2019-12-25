import { Component, OnInit } from "@angular/core";
import { SessionService } from "../shared/session.service";

@Component({
  selector: "app-gc-page",
  templateUrl: "./gc-page.component.html",
  styleUrls: ["./gc-page.component.scss"]
})
export class GcPageComponent implements OnInit {
  inProgress: boolean;
  constructor(private session: SessionService) {}

  ngOnInit() {}
  public get hasAdminRole(): boolean {
    return (
      this.session.getCurrentUser() &&
      (this.session.getCurrentUser().admin_role_in_auth || this.session.getCurrentUser().sysadmin_flag)
    );
  }

  getGcStatus(status: boolean) {
    this.inProgress = status;
  }
}
