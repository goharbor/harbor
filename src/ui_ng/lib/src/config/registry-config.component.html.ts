export const REGISTRY_CONFIG_HTML: string = `
<div>
    <replication-config [(replicationConfig)]="config"></replication-config>
    <system-settings [(systemSettings)]="config"></system-settings>
    <vulnerability-config [(vulnerabilityConfig)]="config"></vulnerability-config>
</div>
`;