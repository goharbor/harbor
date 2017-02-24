import { Component, OnInit } from '@angular/core';

import { Repository } from '../repository/repository';

@Component({
    selector: 'dashboard',
    templateUrl: 'dashboard.component.html'
})
export class DashboardComponent implements OnInit {
    repositories: Repository[];

    ngOnInit(): void {
      this.repositories = [
        { name: 'Ubuntu', version: '14.04',  count: 1 },
        { name: 'MySQL',  version: 'Latest', count: 2 },
        { name: 'Photon', version: '1.0',    count: 3 }
      ];
    }
    
}