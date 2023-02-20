import { Injectable } from '@angular/core';

@Injectable()
export class ImmutableTagService {
    private I18nMap: object = {
        repoMatches: 'MAT',
        repoExcludes: 'EXC',
        matches: 'MAT',
        excludes: 'EXC',
        withLabels: 'WITH',
        withoutLabels: 'WITHOUT',
        none: 'NONE',
        nothing: 'NONE',
    };

    getI18nKey(str: string): string {
        if (this.I18nMap[str.trim()]) {
            return 'IMMUTABLE_TAG.' + this.I18nMap[str.trim()];
        }
        return str;
    }
}
