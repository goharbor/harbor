import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { HttpModule } from '@angular/http';
import { ClarityModule } from '@clr/angular';
import { HarborLibraryModule } from './harbor-library.module';

@NgModule({
    declarations: [],
    imports: [
        BrowserAnimationsModule,
        BrowserModule,
        FormsModule,
        HttpModule,
        ClarityModule.forRoot(),
        HarborLibraryModule.forRoot()
    ],
    providers: [],
    bootstrap: []
})

export class AppModule {
}
