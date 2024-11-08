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
    @Input()
    selectedRow: Artifact[];

    // for tagMode
    @Input()
    selectedTag: string;
    @Input()
    artifact: Artifact;
    @Input()
    accessoryType: string;

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

    getPullCommandForDocker(artifact: Artifact): string {
        return getPullCommandByDigest(
            artifact.type,
            `${this.registryUrl ? this.registryUrl : location.hostname}/${
                this.projectName
            }/${this.repoName}`,
            artifact.digest,
            Clients.DOCKER
        );
    }

    getPullCommandForPadMan(artifact: Artifact): string {
        return getPullCommandByDigest(
            artifact.type,
            `${this.registryUrl ? this.registryUrl : location.hostname}/${
                this.projectName
            }/${this.repoName}`,
            artifact.digest,
            Clients.PODMAN
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

    getPullCommandForDockerByTag(artifact: Artifact): string {
        return getPullCommandByTag(
            artifact.type,
            `${this.registryUrl ? this.registryUrl : location.hostname}/${
                this.projectName
            }/${this.repoName}`,
            this.selectedTag,
            Clients.DOCKER
        );
    }

    getPullCommandForPadManByTag(artifact: Artifact): string {
        return getPullCommandByTag(
            artifact.type,
            `${this.registryUrl ? this.registryUrl : location.hostname}/${
                this.projectName
            }/${this.repoName}`,
            this.selectedTag,
            Clients.PODMAN
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
}
