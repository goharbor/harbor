export const REPLICATION_JOB_TEMPLATE: string = `
<clr-datagrid (clrDgRefresh)="refresh($event)">
    <clr-dg-column>{{'REPLICATION.NAME' | translate}}</clr-dg-column>
    <clr-dg-column>{{'REPLICATION.STATUS' | translate}}</clr-dg-column>
    <clr-dg-column>{{'REPLICATION.OPERATION' | translate}}</clr-dg-column>
    <clr-dg-column>{{'REPLICATION.CREATION_TIME' | translate}}</clr-dg-column>
    <clr-dg-column>{{'REPLICATION.END_TIME' | translate}}</clr-dg-column>
    <clr-dg-column>{{'REPLICATION.LOGS' | translate}}</clr-dg-column>
    <clr-dg-row *clrDgItems="let j of jobs" [clrDgItem]='j'>
        <clr-dg-cell>{{j.repository}}</clr-dg-cell>
        <clr-dg-cell>{{j.status}}</clr-dg-cell>
        <clr-dg-cell>{{j.operation}}</clr-dg-cell>
        <clr-dg-cell>{{j.creation_time}}</clr-dg-cell>
        <clr-dg-cell>{{j.update_time}}</clr-dg-cell>
        <clr-dg-cell>
            <a href="/api/jobs/replication/{{j.id}}/log" target="_BLANK">
                <clr-icon shape="clipboard"></clr-icon>
            </a>
        </clr-dg-cell>
    </clr-dg-row>
    <clr-dg-footer>
        {{ jobs ? jobs.length : 0 }} {{'REPLICATION.ITEMS' | translate}}
        <clr-dg-pagination [clrDgPageSize]="5"></clr-dg-pagination>
    </clr-dg-footer>
</clr-datagrid>`; 