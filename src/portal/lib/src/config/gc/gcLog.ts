export class GcJobData {
    id: number;
    job_name: string;
    job_kind: string;
    schedule: Schedule;
    job_status: string;
    job_uuid: string;
    creation_time: string;
    update_time: string;
    delete: boolean;
}

export class Schedule {
    type: string;
    weekday: number;
    offtime: number;
}
export class GcJobViewModel {
    id: number;
    type: string;
    status: string;
    createTime: Date;
    updateTime: Date;
    details: string;
}

export class WeekDay {
    value: number;
    text: string;
}

export class GcScheduleViewModel {
    type: string;
    weekDay: string;
    dailyTime: string;
}

