(function() {
  'use strict';
  angular
    .module('harbor.app', [
      'ngRoute',
      'harbor.layout.navigation',
      'harbor.layout.repository',
      'harbor.layout.user',
      'harbor.layout.log',
      'harbor.services.user',
      'harbor.services.repository',
      'harbor.services.projectmember',
      'harbor.session',
      'harbor.header',
      'harbor.details',
      'harbor.repository',
      'harbor.projectmember',
      'harbor.user',
      'harbor.log'
    ]);
})();