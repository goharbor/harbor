export const PUSH_IMAGE_HTML: string = `
<div>
    <clr-dropdown [clrMenuPosition]="'bottom-right'">
        <button class="btn btn-link btn-font" clrDropdownToggle (click)="onclick()">
        {{ 'PUSH_IMAGE.TITLE' | translate | uppercase}}
        <clr-icon shape="caret down"></clr-icon>
    </button>
        <div class="dropdown-menu" style="min-width:500px;">
            <div class="commands-container">
                <section>
                    <span><h5 class="h5-override">{{ 'PUSH_IMAGE.TITLE' | translate }}</h5></span>
                    <span>
                      <clr-tooltip [clrTooltipDirection]="'top-right'" [clrTooltipSize]="'md'">
                        <clr-icon shape="info-circle" class="info-tips-icon" size="24"></clr-icon>
                          <clr-tooltip-content>
                            {{ 'PUSH_IMAGE.TOOLTIP' | translate }}
                          </clr-tooltip-content>
                      </clr-tooltip>
                    </span>
                </section>
                <section>
                  <hbr-inline-alert #copyAlert></hbr-inline-alert>
                </section>
                <section>
                    <article class="commands-section">
                        <hbr-copy-input #tagCopy (onCopyError)="onCpError($event)" inputSize="50" headerTitle="{{ 'PUSH_IMAGE.TAG_COMMAND' | translate }}" defaultValue="{{tagCommand}}"></hbr-copy-input>
                    </article>
                    <article class="commands-section">
                        <hbr-copy-input #pushCopy (onCopyError)="onCpError($event)" inputSize="50" headerTitle="{{ 'PUSH_IMAGE.PUSH_COMMAND' | translate }}" defaultValue="{{pushCommand}}"></hbr-copy-input>
                    </article>
                </section>
            </div>
        </div>
    </clr-dropdown>
</div>
`;