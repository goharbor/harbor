import { Component, Input } from '@angular/core';
import {
    AccessoryType,
    ArtifactFront as Artifact,
    ArtifactType,
    Clients,
    getPullCommandByDigest,
    getPullCommandByTag,
    hasPullCommand,
} from '../../../../artifact';
import { getContainerRuntime } from 'src/app/shared/units/shared.utils';
import { MessageHandlerService } from 'src/app/shared/services/message-handler.service';
import { TranslateService } from '@ngx-translate/core';

@Component({
    selector: 'app-pull-command',
    templateUrl: './pull-command.component.html',
    styleUrls: ['./pull-command.component.scss'],
})
export class PullCommandComponent {
    @Input()
    isTagMode: boolean = false; // tagMode is for tag list datagrid,
    @Input()
    projectName: string;
    @Input()
    registryUrl: string;
    @Input()
    repoName: string;

    // for tagMode
    @Input()
    selectedTag: string;
    @Input()
    artifact: Artifact;
    @Input()
    accessoryType: string;

    constructor(
        private msgHandler: MessageHandlerService,
        private translate: TranslateService
    ) {}

    hasPullCommand(artifact: Artifact): boolean {
        return hasPullCommand(artifact);
    }

    isImage(artifact: Artifact): boolean {
        return artifact.type === ArtifactType.IMAGE;
    }

    isCNAB(artifact: Artifact): boolean {
        return artifact.type === ArtifactType.CNAB;
    }

    isChart(artifact: Artifact): boolean {
        return artifact.type === ArtifactType.CHART;
    }

    // get client based on the selected container runtime
    getSelectedClient(): Clients {
        const runtime = getContainerRuntime();
        const client = Object.values(Clients).find(client => client == runtime);
        // return client if match found otherwise return (DOCKER)
        return client ? client : Clients.DOCKER;
    }

    getPullCommandForRuntimeByDigest(artifact: Artifact): string {
        return getPullCommandByDigest(
            artifact.type,
            `${this.registryUrl ? this.registryUrl : location.hostname}/${
                this.projectName
            }/${this.repoName}`,
            artifact.digest,
            this.getSelectedClient()
        );
    }

    getPullCommandForCNAB(artifact: Artifact): string {
        return getPullCommandByDigest(
            artifact.type,
            `${this.registryUrl ? this.registryUrl : location.hostname}/${
                this.projectName
            }/${this.repoName}`,
            artifact.digest,
            Clients.CNAB
        );
    }

    getPullCommandForChart(artifact: Artifact): string {
        return getPullCommandByTag(
            artifact.type,
            `${this.registryUrl ? this.registryUrl : location.hostname}/${
                this.projectName
            }/${this.repoName}`,
            artifact.tags[0].name,
            Clients.CHART
        );
    }

    // For tagMode
    hasPullCommandForTag(artifact): boolean {
        return (
            (artifact?.type === ArtifactType.IMAGE ||
                artifact?.type === ArtifactType.CHART ||
                artifact?.type === ArtifactType.CNAB) &&
            this.accessoryType !== AccessoryType.COSIGN &&
            this.accessoryType !== AccessoryType.NOTATION &&
            this.accessoryType !== AccessoryType.NYDUS
        );
    }

    getPullCommandForRuntimeByTag(artifact: Artifact): string {
        return getPullCommandByTag(
            artifact.type,
            `${this.registryUrl ? this.registryUrl : location.hostname}/${
                this.projectName
            }/${this.repoName}`,
            this.selectedTag,
            this.getSelectedClient()
        );
    }

    getPullCommandForCNABByTag(artifact: Artifact): string {
        return getPullCommandByTag(
            artifact.type,
            `${this.registryUrl ? this.registryUrl : location.hostname}/${
                this.projectName
            }/${this.repoName}`,
            this.selectedTag,
            Clients.CNAB
        );
    }

    getPullCommandForChartByTag(artifact: Artifact): string {
        return getPullCommandByTag(
            artifact.type,
            `${this.registryUrl ? this.registryUrl : location.hostname}/${
                this.projectName
            }/${this.repoName}`,
            this.selectedTag,
            Clients.CHART
        );
    }

    onCpSuccess(copied: string): void {
        // $event is the defaultValue emitted from CopyInputComponent
        this.translate
            .get('REPOSITORY.COPY_SUCCESS', {
                param: copied,
            })
            .subscribe((res: string) => {
                this.msgHandler.showSuccess(res);
            });
    }
}
