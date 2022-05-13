export enum Flatten_Level {
    NO_FLATTING = 0,
    FLATTEN_LEVEl_1 = 1,
    FLATTEN_LEVEl_2 = 2,
    FLATTEN_LEVEl_3 = 3,
    FLATTEN_ALL = -1,
}

export const Flatten_I18n_MAP = {
    [Flatten_Level.NO_FLATTING]: 'REPLICATION.NO_FLATTING',
    [Flatten_Level.FLATTEN_LEVEl_1]: 'REPLICATION.FLATTEN_LEVEL_1',
    [Flatten_Level.FLATTEN_LEVEl_2]: 'REPLICATION.FLATTEN_LEVEL_2',
    [Flatten_Level.FLATTEN_LEVEl_3]: 'REPLICATION.FLATTEN_LEVEL_3',
    [Flatten_Level.FLATTEN_ALL]: 'REPLICATION.FLATTEN_ALL',
};

export enum Decoration {
    MATCHES = 'matches',
    EXCLUDES = 'excludes',
}
export enum BandwidthUnit {
    MB = 'Mbps',
    KB = 'Kbps',
}
export enum ReplicationExecutionFilter {
    TRIGGER = 'trigger',
    STATUS = 'status',
}
