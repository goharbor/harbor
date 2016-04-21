(function() {
  'use strict';
  angular
    .module('harbor.app', [
      'ngRoute',
      'harbor.layout.navigation',
      'harbor.layout.repository',
      'harbor.layout.project.member',
      'harbor.layout.user',
      'harbor.layout.log',
      'harbor.services.user',
      'harbor.services.repository',
      'harbor.services.project.member',
      'harbor.session',
      'harbor.header',
      'harbor.details',
      'harbor.repository',
      'harbor.project.member',
      'harbor.user',
      'harbor.log'
    ]);
})();