import { Injectable } from '@angular/core';
import { from, Subject } from "rxjs";

@Injectable({
  providedIn: "root"
})
export class EventService {

  private listeners = {};
  private eventsSubject = new Subject();
  private events = from(this.eventsSubject);

  constructor() {
    this.events.subscribe(
      ({name, args}) => {
        if (this.listeners[name]) {
          for (let listener of this.listeners[name]) {
            listener(...args);
          }
        }
      });
  }
  subscribe(name: string, listener): any {
    if (!this.listeners[name]) {
      this.listeners[name] = [];
    }
    this.listeners[name].push(listener);
    return {
      unsubscribe: () => {
        this.doUnsubscribe(name, listener);
      }
    };
  }
  doUnsubscribe(name, listener) {
    this.listeners[name] = this.listeners[name].filter((v) => {
      return v !== listener;
    });
  }
  unsubscribe(name, listener?) {
    if (this.listeners[name]) {
      if (!listener) {
        this.listeners[name] = [];
      } else {
        this.doUnsubscribe(name, listener);
      }
    }
  }
  publish(name, ...args) {
    this.eventsSubject.next({
      name,
      args
    });
  }
}


export enum HarborEvent {
  SCROLL = 'scroll',
  SCROLL_TO_POSITION = 'scrollToPosition'
}
