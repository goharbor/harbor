import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of } from 'rxjs';
import { RepositoryService as NewRepositoryService } from '../../../../../ng-swagger-gen/services/repository.service';
import { RepositoryGridviewComponent } from './repository-gridview.component';
import {
    ProjectDefaultService,
    ProjectService,
    SystemInfo,
    SystemInfoService,
    UserPermissionService,
} from '../../../shared/services';
import { delay } from 'rxjs/operators';
import { ErrorHandler } from '../../../shared/units/error-handler';
import { ActivatedRoute } from '@angular/router';
import { Repository as NewRepository } from '../../../../../ng-swagger-gen/models/repository';
import { SharedTestingModule } from '../../../shared/shared.module';
import { GridViewComponent } from './gridview/grid-view.component';

describe('RepositoryComponentGridview (inline template)', () => {
    let compRepo: RepositoryGridviewComponent;
    let fixtureRepo: ComponentFixture<RepositoryGridviewComponent>;
    let mockSystemInfo: SystemInfo = {
        with_notary: true,
        with_admiral: false,
        admiral_endpoint: 'NA',
        auth_mode: 'db_auth',
        registry_url: '10.112.122.56',
        project_creation_restriction: 'everyone',
        self_registration: true,
        has_ca_root: false,
        harbor_version: 'v1.1.1-rc1-160-g565110d',
    };
    let mockRepoData: NewRepository[] = [
        {
            id: 1,
            name: 'library/busybox',
            project_id: 1,
            description: 'asdfsadf',
            pull_count: 0,
            artifact_count: 1,
        },
        {
            id: 2,
            name: 'library/nginx',
            project_id: 1,
            description: 'asdf',
            pull_count: 0,
            artifact_count: 1,
        },
    ];
    let mockRepoNginxData: NewRepository[] = [
        {
            id: 2,
            name: 'library/nginx',
            project_id: 1,
            description: 'asdf',
            pull_count: 0,
            artifact_count: 1,
        },
    ];

    let mockRepo: NewRepository[] = mockRepoData;
    let mockNginxRepo: NewRepository[] = mockRepoNginxData;
    const fakedErrorHandler = {
        error() {
            return undefined;
        },
    };
    const fakedSystemInfoService = {
        getSystemInfo() {
            return of(mockSystemInfo);
        },
    };
    const fakedRepositoryService = {
        listRepositoriesResponse(
            params: NewRepositoryService.ListRepositoriesParams
        ) {
            if (params.q === encodeURIComponent(`name=~nginx`)) {
                return of({ headers: new Map(), body: mockNginxRepo });
            }
            return of({ headers: new Map(), body: mockRepo }).pipe(delay(0));
        },
    };
    const fakedUserPermissionService = {
        getPermission() {
            return of(true);
        },
    };
    const fakedActivatedRoute = {
        snapshot: {
            parent: {
                parent: {
                    params: {
                        id: '1',
                    },
                },
            },
        },
    };
    beforeEach(() => {
        let store = {};
        spyOn(localStorage, 'getItem').and.callFake(key => {
            return store[key];
        });
        spyOn(localStorage, 'setItem').and.callFake((key, value) => {
            return (store[key] = value + '');
        });
        spyOn(localStorage, 'clear').and.callFake(() => {
            store = {};
        });
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [RepositoryGridviewComponent, GridViewComponent],
            providers: [
                { provide: ActivatedRoute, useValue: fakedActivatedRoute },
                { provide: ErrorHandler, useValue: fakedErrorHandler },
                {
                    provide: NewRepositoryService,
                    useValue: fakedRepositoryService,
                },
                { provide: ProjectService, useClass: ProjectDefaultService },
                {
                    provide: SystemInfoService,
                    useValue: fakedSystemInfoService,
                },
                {
                    provide: UserPermissionService,
                    useValue: fakedUserPermissionService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(async () => {
        fixtureRepo = TestBed.createComponent(RepositoryGridviewComponent);
        compRepo = fixtureRepo.componentInstance;
        compRepo.projectId = 1;
        compRepo.mode = '';
        compRepo.hasProjectAdminRole = true;
        compRepo.isCardView = false;
        fixtureRepo.detectChanges();
    });
    let originalTimeout;
    beforeEach(function () {
        originalTimeout = jasmine.DEFAULT_TIMEOUT_INTERVAL;
        jasmine.DEFAULT_TIMEOUT_INTERVAL = 100000;
    });
    afterEach(function () {
        jasmine.DEFAULT_TIMEOUT_INTERVAL = originalTimeout;
    });

    it('should create', () => {
        expect(compRepo).toBeTruthy();
    });
    it('should be card view', async () => {
        const cardViewButton =
            fixtureRepo.nativeElement.querySelector('.card-btn');
        cardViewButton.click();
        cardViewButton.dispatchEvent(new Event('click'));
        fixtureRepo.detectChanges();
        await fixtureRepo.whenStable();
        const cordView =
            fixtureRepo.nativeElement.querySelector('hbr-gridview');
        expect(cordView).toBeTruthy();
    });
    it('should be list view', async () => {
        const listViewButton =
            fixtureRepo.nativeElement.querySelector('.list-btn');
        listViewButton.click();
        listViewButton.dispatchEvent(new Event('click'));
        fixtureRepo.detectChanges();
        await fixtureRepo.whenStable();
        const listView =
            fixtureRepo.nativeElement.querySelector('clr-datagrid');
        expect(listView).toBeTruthy();
    });
});
