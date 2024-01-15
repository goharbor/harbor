import {
    Component,
    EventEmitter,
    Input,
    OnChanges,
    OnDestroy,
    OnInit,
    Output,
    SimpleChanges,
} from '@angular/core';
import { Label } from '../../../../../ng-swagger-gen/models/label';
import { forkJoin, Observable, Subject, Subscription } from 'rxjs';
import { LabelService } from '../../../../../ng-swagger-gen/services/label.service';
import {
    debounceTime,
    distinctUntilChanged,
    filter,
    finalize,
    map,
    switchMap,
} from 'rxjs/operators';
import { Router } from '@angular/router';

const GLOBAL: string = 'g';
const PROJECT: string = 'p';
const PAGE_SIZE: number = 50;

@Component({
    selector: 'app-label-selector',
    templateUrl: './label-selector.component.html',
    styleUrls: ['./label-selector.component.scss'],
})
export class LabelSelectorComponent implements OnInit, OnChanges, OnDestroy {
    @Input()
    usedInDropdown: boolean = false;
    @Input()
    ownedLabels: Label[] = [];
    @Input()
    width: number = 180; // unit px
    @Input()
    scope: string = GLOBAL; // 'g' for global and 'p' for project, default 'g'
    @Input()
    projectId: number; // if scope = 'p', projectId is required
    @Input()
    dropdownOpened: boolean; // parent component opened status
    candidateLabels: Label[] = [];
    searchValue: string;
    loading: boolean = false;
    @Output()
    clickLabel = new EventEmitter<{
        label: Label;
        isAdd: boolean;
    }>();
    private _searchSubject = new Subject<string>();
    private _subSearch: Subscription;
    constructor(private labelService: LabelService, private router: Router) {
        if (!this._subSearch) {
            this._subSearch = this._searchSubject
                .pipe(
                    debounceTime(500),
                    distinctUntilChanged(),
                    filter(labelName => {
                        if (!labelName) {
                            this.initCandidateLabel();
                        }
                        return !!labelName;
                    }),
                    switchMap(labelName => {
                        return this.getLabelObservable(labelName);
                    })
                )
                .subscribe(res => {
                    this.candidateLabels = res;
                });
        }
    }

    ngOnInit(): void {
        this.checkProjectId();
        this.initCandidateLabel();
    }

    initCandidateLabel() {
        // Place the owned label at the top of the array then remove duplicates
        const Obs: Observable<Label[]>[] = [];
        if (this.ownedLabels?.length) {
            const projectLabelIds: number[] = [];
            const globalLabelIds: number[] = [];
            this.ownedLabels?.forEach(item => {
                if (item.scope === PROJECT) {
                    projectLabelIds.push(item.id);
                }
                if (item.scope === GLOBAL) {
                    globalLabelIds.push(item.id);
                }
            });
            if (projectLabelIds?.length) {
                Obs.push(
                    this.labelService.ListLabels({
                        page: 1,
                        pageSize: PAGE_SIZE,
                        scope: PROJECT,
                        projectId: this.projectId,
                        q: encodeURIComponent(
                            `id={${projectLabelIds.join(' ')}}`
                        ),
                    })
                );
            }
            if (globalLabelIds?.length) {
                Obs.push(
                    this.labelService.ListLabels({
                        page: 1,
                        pageSize: PAGE_SIZE,
                        scope: GLOBAL,
                        q: encodeURIComponent(
                            `id={${globalLabelIds.join(' ')}}`
                        ),
                    })
                );
            }
        }
        Obs.push(this.getLabelObservable(''));
        forkJoin(Obs)
            .pipe(
                map(result => [].concat.apply([], result)),
                map((result: Label[]) => {
                    return result.filter(
                        (v, i, a) => a.findIndex(v2 => v2.id === v.id) === i
                    );
                })
            )
            .subscribe(res => {
                this.candidateLabels = res;
            });
    }

    ngOnChanges(changes: SimpleChanges): void {
        this.checkProjectId();
    }
    ngOnDestroy() {
        if (this._subSearch) {
            this._subSearch.unsubscribe();
            this._subSearch = null;
        }
    }

    checkProjectId() {
        if (this.scope === PROJECT && !this.projectId) {
            throw new Error('Attribute [projectId] is required');
        }
    }

    search() {
        this._searchSubject.next(this.searchValue);
    }

    selectLabel(label: Label) {
        this.clickLabel.emit({ label: label, isAdd: !this.isSelect(label) });
    }
    isSelect(label: Label): boolean {
        if (this.ownedLabels?.length) {
            return this.ownedLabels.some(item => {
                return item.id === label.id && this.dropdownOpened;
            });
        }
        return false;
    }
    goToLabelPage() {
        if (this.scope === PROJECT) {
            this.router.navigate([
                'harbor',
                'projects',
                this.projectId,
                'labels',
            ]);
        } else {
            this.router.navigate(['harbor', 'labels']);
        }
    }

    getLabelObservable(labelName: string): Observable<Label[]> {
        this.loading = true;
        if (this.scope === PROJECT) {
            return forkJoin([
                this.labelService.ListLabels({
                    page: 1,
                    pageSize: PAGE_SIZE,
                    scope: PROJECT,
                    projectId: this.projectId,
                    q: labelName
                        ? encodeURIComponent(`name=~${labelName}`)
                        : null,
                }),
                this.labelService.ListLabels({
                    page: 1,
                    pageSize: PAGE_SIZE,
                    scope: GLOBAL,
                    q: labelName
                        ? encodeURIComponent(`name=~${labelName}`)
                        : null,
                }),
            ]).pipe(
                map(result => [].concat.apply([], result)),
                finalize(() => (this.loading = false))
            );
        }
        return this.labelService
            .ListLabels({
                page: 1,
                pageSize: PAGE_SIZE,
                scope: GLOBAL,
                q: labelName ? encodeURIComponent(`name=~${labelName}`) : null,
            })
            .pipe(finalize(() => (this.loading = false)));
    }
}
