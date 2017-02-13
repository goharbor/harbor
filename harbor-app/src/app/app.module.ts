import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { HttpModule } from '@angular/http';
import { ClarityModule } from 'clarity-angular';
import { AppComponent } from './app.component';
import { AccountModule } from './account/account.module';
import { HarborShellModule } from './harbor-shell/harbor.shell.module';
import { HarborRoutingModule } from './harbor-routing.module';
import { SharedModule } from './shared.module';

@NgModule({
    declarations: [
        AppComponent,
    ],
    imports: [
        SharedModule,
        AccountModule,
        HarborShellModule,
        HarborRoutingModule
    ],
    providers: [],
    bootstrap: [ AppComponent ]
})
export class AppModule {
}
