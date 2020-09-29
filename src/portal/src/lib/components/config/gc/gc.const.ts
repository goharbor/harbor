import { OriginCron } from "../../../services";


export const SCHEDULE_TYPE_NONE = "None";

export const ONE_MINITUE = 60000;
export const THREE_SECONDS = 3000;

export interface GCSchedule {
  schedule?: OriginCron;
  parameters?: {[key: string]: any};
  id?: number;
  job_name?: string;
  job_kind?: string;
  job_parameters?: string;
  job_status?: string;
  deleted?: boolean;
  creation_time?: Date;
  update_time?: Date;
}





