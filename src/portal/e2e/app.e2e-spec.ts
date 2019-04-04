// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import {ClaritySeedAppHome} from './app.po';

fdescribe('harbor-portal app', function () {

  let expectedMsg: string = 'This is a Clarity seed application. This is the default page that loads for the application.';

  let page: ClaritySeedAppHome;

  beforeEach(() => {
    page = new ClaritySeedAppHome();
  });

  it('should display: ' + expectedMsg, () => {
    page.navigateTo();
    page.getParagraphText().then(res => {
      expect(res).toEqual(expectedMsg);
    });
  });
});
