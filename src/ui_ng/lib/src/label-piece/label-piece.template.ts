/**
 * Define template resources for filter component
 */

export const LABEL_PIEICE_TEMPLATE: string = `
<label class="label" [ngStyle]="{'background-color': label.color}" [style.max-width.px]="labelWidth">
    <clr-icon *ngIf="label.scope=='p'" shape="organization"></clr-icon>
    <clr-icon *ngIf="label.scope=='g'" shape="administrator"></clr-icon>
     {{label.name}}
</label>
`;

export const LABEL_PIEICE_STYLES: string = `
   .label{border: none; color:#222;
      display: inline-block;
      justify-content: flex-start;
      overflow: hidden;
      text-overflow: ellipsis;
      line-height: .875rem;}
   .label clr-icon{ margin-right: 3px;}
   .btn-group .dropdown-menu clr-icon{display:block;}
`;