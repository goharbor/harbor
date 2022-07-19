import { ImmutableTagService } from './immutable-tag.service';
import { TestBed, inject } from '@angular/core/testing';
import {
    HttpClientTestingModule,
    HttpTestingController,
} from '@angular/common/http/testing';

describe('ImmutableTagService', () => {
    beforeEach(() =>
        TestBed.configureTestingModule({
            providers: [ImmutableTagService],
            imports: [HttpClientTestingModule],
        })
    );

    it('should be created', () => {
        const service: ImmutableTagService = TestBed.get(ImmutableTagService);
        expect(service).toBeTruthy();
    });
    it('should get rules', inject(
        [HttpTestingController, ImmutableTagService],
        (
            httpMock: HttpTestingController,
            immutableTagService: ImmutableTagService
        ) => {
            const mockRules = [
                {
                    id: 1,
                    project_id: 1,
                    disabled: false,
                    priority: 0,
                    action: 'immutable',
                    template: 'immutable_template',
                    tag_selectors: [
                        {
                            kind: 'doublestar',
                            decoration: 'matches',
                            pattern: '**',
                        },
                    ],
                    scope_selectors: {
                        repository: [
                            {
                                kind: 'doublestar',
                                decoration: 'repoMatches',
                                pattern: '**',
                            },
                        ],
                    },
                },
                {
                    id: 2,
                    project_id: 1,
                    disabled: false,
                    priority: 0,
                    action: 'immutable',
                    template: 'immutable_template',
                    tag_selectors: [
                        {
                            kind: 'doublestar',
                            decoration: 'matches',
                            pattern: '44',
                        },
                    ],
                    scope_selectors: {
                        repository: [
                            {
                                kind: 'doublestar',
                                decoration: 'repoMatches',
                                pattern: '**',
                            },
                        ],
                    },
                },
                {
                    id: 3,
                    project_id: 1,
                    disabled: false,
                    priority: 0,
                    action: 'immutable',
                    template: 'immutable_template',
                    tag_selectors: [
                        {
                            kind: 'doublestar',
                            decoration: 'matches',
                            pattern: '555',
                        },
                    ],
                    scope_selectors: {
                        repository: [
                            {
                                kind: 'doublestar',
                                decoration: 'repoMatches',
                                pattern: '**',
                            },
                        ],
                    },
                },
                {
                    id: 4,
                    project_id: 1,
                    disabled: false,
                    priority: 0,
                    action: 'immutable',
                    template: 'immutable_template',
                    tag_selectors: [
                        {
                            kind: 'doublestar',
                            decoration: 'matches',
                            pattern: 'fff**',
                        },
                    ],
                    scope_selectors: {
                        repository: [
                            {
                                kind: 'doublestar',
                                decoration: 'repoMatches',
                                pattern: '**ggg',
                            },
                        ],
                    },
                },
            ];

            immutableTagService.getRules(1).subscribe(res => {
                expect(res).toEqual(mockRules);
            });
        }
    ));
});
