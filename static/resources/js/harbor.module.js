/*
    Copyright (c) 2016 VMware, Inc. All Rights Reserved.
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
        
        http://www.apache.org/licenses/LICENSE-2.0
        
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/
(function() {
  'use strict';
  angular
    .module('harbor.app', [
      'ngMessages',
      'ngCookies',
      'harbor.session',
      'harbor.layout.element.height',
      'harbor.layout.header',
      'harbor.layout.footer',
      'harbor.layout.navigation',
      'harbor.layout.sign.up',
      'harbor.layout.add.new',
      'harbor.layout.account.setting',
      'harbor.layout.change.password',
      'harbor.layout.forgot.password',
      'harbor.layout.reset.password',
      'harbor.layout.index',
      'harbor.layout.dashboard',
      'harbor.layout.project',
      'harbor.layout.admin.option',
      'harbor.layout.search',
      'harbor.services.i18n',
      'harbor.services.project',
      'harbor.services.user',
      'harbor.services.repository',
      'harbor.services.project.member',
      'harbor.services.replication.policy',
      'harbor.services.replication.job',
      'harbor.services.destination',
      'harbor.summary',
      'harbor.user.log',
      'harbor.top.repository',
      'harbor.optional.menu',
      'harbor.modal.dialog',
      'harbor.sign.in',
      'harbor.search',
      'harbor.project',
      'harbor.details',
      'harbor.repository',
      'harbor.project.member',
      'harbor.user',
      'harbor.log',
      'harbor.validator',
      'harbor.replication',
      'harbor.system.management',
      'harbor.loading.progress',
      'harbor.inline.help',
      'harbor.dismissable.alerts',
      'harbor.paginator'
    ]);
})();
