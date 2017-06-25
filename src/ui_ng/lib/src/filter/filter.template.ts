/**
 * Define template resources for filter component
 */

export const FILTER_TEMPLATE: string = `
<span>
    <clr-icon shape="search" size="20" class="search-btn" [class.filter-icon]="isShowSearchBox" (click)="onClick()"></clr-icon>
    <input [hidden]="!isShowSearchBox" type="text" style="padding-left: 15px;" (keyup)="valueChange()" placeholder="{{placeHolder}}" [(ngModel)]="currentValue"/>
    <span class="filter-divider" *ngIf="withDivider"></span>
</span>
`;

export const FILTER_STYLES: string = `
.filter-icon {
    position: relative;
    right: -12px;
}

.filter-divider {
    display: inline-block;
    height: 16px;
    width: 2px;
    background-color: #cccccc;
    padding-top: 12px;
    padding-bottom: 12px;
    position: relative;
    top: 9px;
    margin-right: 6px;
    margin-left: 6px;
}

.search-btn {
    cursor: pointer;
}

.search-btn:hover {
    color: #007CBB;
}
`;