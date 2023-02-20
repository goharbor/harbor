import { Component, OnInit, ViewChild } from '@angular/core';
import { dbEncodeURIComponent } from '../../../../../../../../shared/units/utils';
import { finalize } from 'rxjs/operators';
import { ArtifactService } from 'ng-swagger-gen/services/artifact.service';
import { ImageNameInputComponent } from '../../../../../../../../shared/components/image-name-input/image-name-input.component';
import { Project } from '../../../../../../project';
import { ActivatedRoute } from '@angular/router';
import { ErrorHandler } from '../../../../../../../../shared/units/error-handler';
import { TranslateService } from '@ngx-translate/core';

@Component({
    selector: 'app-copy-artifact',
    templateUrl: './copy-artifact.component.html',
    styleUrls: ['./copy-artifact.component.scss'],
})
export class CopyArtifactComponent implements OnInit {
    retagDialogOpened: boolean = false;
    projectName: string;
    repoName: string;
    @ViewChild('imageNameInput')
    imageNameInput: ImageNameInputComponent;
    digest: string;
    constructor(
        private activatedRoute: ActivatedRoute,
        private artifactService: ArtifactService,
        private errorHandlerService: ErrorHandler,
        private translateService: TranslateService
    ) {}

    ngOnInit(): void {
        const resolverData = this.activatedRoute.snapshot?.parent?.parent?.data;
        if (resolverData) {
            this.projectName = (<Project>resolverData['projectResolver']).name;
        }
        this.repoName = this.activatedRoute.snapshot?.parent?.params['repo'];
    }

    onRetag() {
        let params: ArtifactService.CopyArtifactParams = {
            projectName: this.imageNameInput.projectName.value,
            repositoryName: dbEncodeURIComponent(
                this.imageNameInput.repoName.value
            ),
            from: `${this.projectName}/${this.repoName}@${this.digest}`,
        };
        this.artifactService
            .CopyArtifact(params)
            .pipe(
                finalize(() => {
                    this.imageNameInput.form.reset();
                    this.retagDialogOpened = false;
                })
            )
            .subscribe({
                next: response => {
                    this.translateService
                        .get('RETAG.MSG_SUCCESS')
                        .subscribe((res: string) => {
                            this.errorHandlerService.info(res);
                        });
                },
                error: error => {
                    this.errorHandlerService.error(error);
                },
            });
    }
    retag(digest: string) {
        this.retagDialogOpened = true;
        this.imageNameInput.imageNameForm.reset({
            repoName: this.repoName,
            projectName: null,
        });
        this.digest = digest;
        this.imageNameInput.notExist = false;
    }
}
