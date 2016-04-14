(function() {
  'use strict';
  angular
    .module('harbor.app', [
      'harbor.services.user',
      'harbor.session',
      'harbor.header',
      'harbor.details'
    ]);
})();