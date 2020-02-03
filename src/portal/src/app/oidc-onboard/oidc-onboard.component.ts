import { Router, ActivatedRoute } from '@angular/router';
import { Component, OnInit } from '@angular/core';
import { OidcOnboardService } from './oidc-onboard.service';
import { FormControl } from '@angular/forms';
import { CommonRoutes } from "../../lib/entities/shared.const";
import { errorHandler } from "../../lib/utils/shared/shared.utils";

@Component({
  selector: 'app-oidc-onboard',
  templateUrl: './oidc-onboard.component.html',
  styleUrls: ['./oidc-onboard.component.scss']
})
export class OidcOnboardComponent implements OnInit {
  url: string;
  errorMessage: string = '';
  oidcUsername = new FormControl('');
  errorOpen: boolean = false;
  constructor(
    private oidcOnboardService: OidcOnboardService,
    private router: Router,
    private route: ActivatedRoute,
  ) { }

  ngOnInit() {
    this.route.queryParams
      .subscribe(params => {
        this.oidcUsername.setValue(params["username"] || "");
      });
  }
  clickSaveBtn(): void {
    this.oidcOnboardService.oidcSave({ username: this.oidcUsername.value }).subscribe(res => {
      this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
    }
      , error => {
        this.errorMessage = errorHandler(error);
        this.errorOpen = true;
      });
  }
  emptyErrorMessage() {
    this.errorOpen = false;
  }
  backHarborPage() {
    this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
  }
}
