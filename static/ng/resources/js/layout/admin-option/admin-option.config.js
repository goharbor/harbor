(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.admin.option')
    .config(routeConfig);
    
  function routeConfig($routeProvider) {
    $routeProvider
      .when('/all_user', {
        'templateUrl': '/static/ng/resources/js/layout/user/user.controller.html',
        'controller': 'UserController', 
        'controllerAs': 'vm'
      })
      .when('/system_management', {
        'templateUrl': '/static/ng/resources/js/layout/system-management/system-management.controller.html',
        'controller': 'SystemManagementController',
        'controllerAs': 'vm'
      })
      .otherwise({
        'redirectTo': '/'
      });
  }
  
})();