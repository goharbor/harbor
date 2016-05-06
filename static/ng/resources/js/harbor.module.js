(function() {
  'use strict';
  angular
    .module('harbor.app', [
      'ngRoute',
      'harbor.layout.navigation',
      'harbor.layout.sign.up',
      'harbor.layout.index',
      'harbor.layout.project',
      'harbor.layout.repository',
      'harbor.layout.project.member',
      'harbor.layout.user',
      'harbor.layout.log',
      'harbor.services.project',
      'harbor.services.user',
      'harbor.services.repository',
      'harbor.services.project.member',
      'harbor.session',
      'harbor.header',
      'harbor.sign.in',
      'harbor.search',
      'harbor.project',
      'harbor.details',
      'harbor.repository',
      'harbor.project.member',
      'harbor.user',
      'harbor.log'
    ]);
})();