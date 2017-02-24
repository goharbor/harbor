import { browser, element, by } from 'protractor';


export class ClaritySeedAppHome {

  navigateTo() {
    return browser.get('/');
  }

  getParagraphText() {
    return element(by.css('my-app p')).getText();
  }
}
