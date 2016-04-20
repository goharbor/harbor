(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .config(routeConfig);
  
  function routeConfig($routeProvider) {
    $routeProvider
      .when('/repositories', {
        templateUrl: '/static/ng/resources/js/layout/repository/repository.controller.html',
        controller: 'RepositoryController',
        controllerAs: 'vm'
       })
      .when('/users', {
        templateUrl: '/static/ng/resources/js/layout/user/user.controller.html',
        controller: 'UserController',
        controllerAs: 'vm'
      })
      .when('/logs', {
        templateUrl: '/static/ng/resources/js/layout/log/log.controller.html',
        controller: 'LogController',
        controllerAs: 'vm'
      })
      .otherwise({
        redirectTo: '/'
      });
  }
  
})();