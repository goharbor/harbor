import './polyfills.ts';

import { platformBrowserDynamic } from '@angular/platform-browser-dynamic';
import { enableProdMode } from '@angular/core';
// import { environment } from './environments/environment';
import { AppModuleNgFactory } from '../aot/src/app/app.module.ngfactory';


enableProdMode();

platformBrowserDynamic().bootstrapModuleFactory(AppModuleNgFactory);
