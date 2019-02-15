
import { Injectable } from '@angular/core';

const ONE_HOUR_SECONDS: number = 3600;
const ONE_DAY_SECONDS: number = 24 * ONE_HOUR_SECONDS;

@Injectable()
export class GcUtility {
    private _localTime: Date = new Date();
    public  getOffTime(v: string) {
        let values: string[] = v.split(":");
        if (!values || values.length !== 2) {
          return;
        }
        let hours: number = +values[0];
        let minutes: number = +values[1];
        // Convert to UTC time
        let timezoneOffset: number = this._localTime.getTimezoneOffset();
        let utcTimes: number = hours * ONE_HOUR_SECONDS + minutes * 60;
        utcTimes += timezoneOffset * 60;
        if (utcTimes < 0) {
          utcTimes += ONE_DAY_SECONDS;
        }
        if (utcTimes >= ONE_DAY_SECONDS) {
          utcTimes -= ONE_DAY_SECONDS;
        }
        return utcTimes;
    }

    public getDailyTime(v: number ) {
    let timeOffset: number = 0; // seconds
    timeOffset = + v;
    // Convert to current time
    let timezoneOffset: number = this._localTime.getTimezoneOffset();
    // Local time
    timeOffset = timeOffset - timezoneOffset * 60;
    if (timeOffset < 0) {
        timeOffset = timeOffset + ONE_DAY_SECONDS;
    }

    if (timeOffset >= ONE_DAY_SECONDS) {
        timeOffset -= ONE_DAY_SECONDS;
    }

    // To time string
    let hours: number = Math.floor(timeOffset / ONE_HOUR_SECONDS);
    let minutes: number = Math.floor((timeOffset - hours * ONE_HOUR_SECONDS) / 60);

    let timeStr: string = "" + hours;
    if (hours < 10) {
        timeStr = "0" + timeStr;
    }
    if (minutes < 10) {
        timeStr += ":0";
    } else {
        timeStr += ":";
    }
    timeStr += minutes;

    return timeStr;
    }
}
