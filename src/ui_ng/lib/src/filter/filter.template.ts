/**
 * Define template resources for filter component
 */

export const FILTER_TEMPLATE: string = `
<span>
    <clr-icon shape="filter" size="12" class="is-solid filter-icon"></clr-icon>
    <input type="text" style="padding-left: 15px;" (keyup)="valueChange()" placeholder="{{placeHolder}}" [(ngModel)]="currentValue"/>
</span>
`;

export const FILTER_STYLES: string = `
.filter-icon {
    position: relative;
    right: -12px;
}
`;