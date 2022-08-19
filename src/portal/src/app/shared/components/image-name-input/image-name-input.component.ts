import { Component, OnDestroy, OnInit } from '@angular/core';
import { Subject } from 'rxjs';
import { debounceTime, finalize, switchMap } from 'rxjs/operators';
import {
    AbstractControl,
    UntypedFormBuilder,
    UntypedFormGroup,
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
    selectedProjectList: Project[] = [];
    proNameChecker: Subject<string> = new Subject<string>();
    imageNameForm: UntypedFormGroup;
    notExist: boolean = false;
    checkingName: boolean = false;
    public project: string;
    public repo: string;
    public tag: string;

    constructor(
        private fb: UntypedFormBuilder,
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
                    this.notExist = false;
                    this.checkingName = true;
                    this.selectedProjectList = [];
                    return this.proService
                        .listProjects({
                            name: name,
                            page: 1,
                            pageSize: 10,
                        })
                        .pipe(finalize(() => (this.checkingName = false)));
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
                        this.notExist = !exist;
                    } else {
                        this.notExist = true;
                    }
                },
                (error: any) => {
                    this.errorHandler.error(error);
                    this.notExist = true;
                }
            );
    }

    validateProjectName(): void {
        let cont = this.imageNameForm.controls['projectName'];
        if (cont && cont.valid) {
            this.proNameChecker.next(cont.value);
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
        this.notExist = false;
    }
}
