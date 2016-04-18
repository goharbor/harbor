(function() {
  'use strict';
  angular
    .module('harbor.app', [
      'ngRoute',
      'harbor.services.user',
      'harbor.services.repository',
      'harbor.session',
      'harbor.header',
      'harbor.details',
      'harbor.repository',
      'harbor.user',
      'harbor.log'
    ]);
})();