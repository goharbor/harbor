import { NgModule } from '@angular/core';
import { InjectionToken } from '@angular/core';

export const WINDOW = new InjectionToken<Window>('WindowToken');

export function _window(): Window {
    return window;
}

@NgModule({
    providers: [
        {
            provide: WINDOW,
            useFactory: _window,
        },
    ],
})
export class WindowTokenModule {}
