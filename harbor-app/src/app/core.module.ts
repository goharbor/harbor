import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { HttpModule } from '@angular/http';
import { ClarityModule } from 'clarity-angular';
import { AppComponent } from './app.component';
import { AccountModule } from './account/account.module';

@NgModule({
  imports: [
      BrowserModule,
      FormsModule,
      HttpModule,
      ClarityModule
  ],
  exports: [
      BrowserModule,
      FormsModule,
      HttpModule,
      ClarityModule
  ]
})
export class CoreModule {
}
