import { Component, OnDestroy, OnInit } from '@angular/core';
import { Subject } from 'rxjs';
import { debounceTime, switchMap } from 'rxjs/operators';
import {
    AbstractControl,
    FormBuilder,
    FormGroup,
    Validators,
} from '@angular/forms';
import { ErrorHandler } from '../../units/error-handler';
import { ProjectService } from 'ng-swagger-gen/services/project.service';
import { Project } from 'ng-swagger-gen/models/project';

@Component({
    selector: 'hbr-image-name-input',
    templateUrl: './image-name-input.component.html',
    styleUrls: ['./image-name-input.component.scss'],
})
export class ImageNameInputComponent implements OnInit, OnDestroy {
    noProjectInfo = '';
    selectedProjectList: Project[] = [];
    proNameChecker: Subject<string> = new Subject<string>();
    imageNameForm: FormGroup;
    public project: string;
    public repo: string;
    public tag: string;

    constructor(
        private fb: FormBuilder,
        private errorHandler: ErrorHandler,
        private proService: ProjectService
    ) {
        this.imageNameForm = this.fb.group({
            projectName: [
                '',
                Validators.compose([
                    Validators.minLength(2),
                    Validators.required,
                    Validators.pattern('^[a-z0-9]+(?:[._-][a-z0-9]+)*$'),
                ]),
            ],
            repoName: [
                '',
                Validators.compose([
                    Validators.required,
                    Validators.maxLength(256),
                    Validators.pattern(
                        '^[a-z0-9]+(?:[._-][a-z0-9]+)*(/[a-z0-9]+(?:[._-][a-z0-9]+)*)*'
                    ),
                ]),
            ],
        });
    }

    ngOnInit(): void {
        this.proNameChecker
            .pipe(debounceTime(200))
            .pipe(
                switchMap(name => {
                    this.noProjectInfo = '';
                    this.selectedProjectList = [];
                    return this.proService.listProjects({
                        name: name,
                        page: 1,
                        pageSize: 10,
                    });
                })
            )
            .subscribe(
                response => {
                    if (response) {
                        this.selectedProjectList = response;
                        // if input project name exist in the project list
                        let exist = response.find(
                            (data: Project) =>
                                data.name ===
                                this.imageNameForm.controls['projectName'].value
                        );
                        if (!exist) {
                            this.noProjectInfo = 'REPLICATION.NO_PROJECT_INFO';
                        } else {
                            this.noProjectInfo = '';
                        }
                    } else {
                        this.noProjectInfo = 'REPLICATION.NO_PROJECT_INFO';
                    }
                },
                (error: any) => {
                    this.errorHandler.error(error);
                    this.noProjectInfo = 'REPLICATION.NO_PROJECT_INFO';
                }
            );
    }

    validateProjectName(): void {
        let cont = this.imageNameForm.controls['projectName'];
        if (cont && cont.valid) {
            this.proNameChecker.next(cont.value);
        } else {
            this.noProjectInfo = 'PROJECT.NAME_TOOLTIP';
        }
    }

    get form(): AbstractControl {
        return this.imageNameForm;
    }

    get projectName(): AbstractControl {
        return this.imageNameForm.get('projectName');
    }

    get repoName(): AbstractControl {
        return this.imageNameForm.get('repoName');
    }

    ngOnDestroy(): void {
        if (this.proNameChecker) {
            this.proNameChecker.unsubscribe();
        }
    }

    leaveProjectInput(): void {
        this.selectedProjectList = [];
    }

    selectedProjectName(projectName: string) {
        this.imageNameForm.controls['projectName'].setValue(projectName);
        this.selectedProjectList = [];
        this.noProjectInfo = '';
    }
}
