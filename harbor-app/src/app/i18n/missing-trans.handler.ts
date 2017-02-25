import {MissingTranslationHandler, MissingTranslationHandlerParams} from '@ngx-translate/core';

export class MyMissingTranslationHandler implements MissingTranslationHandler {
    handle(params: MissingTranslationHandlerParams) {
        const missingText = "{Miss Harbor Text}";
        return missingText;
    }
}