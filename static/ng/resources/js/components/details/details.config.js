(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .config(routeConfig)
    .filter('name', nameFilter);
    
  function routeConfig($routeProvider) {
    $routeProvider
      .when('/repositories', {
        templateUrl: '/static/ng/resources/js/layout/repository/repository.controller.html',
        controller: 'RepositoryController',
        controllerAs: 'vm'
       })
      .when('/users', {
        templateUrl: '/static/ng/resources/js/layout/project-member/project-member.controller.html',
        controller: 'ProjectMemberController',
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
  
  function nameFilter() {
   
    return filter;

    function filter(input, filterInput, key) {
      input = input || '';
      var filteredResults = [];
 
      if (filterInput != '') {
        for(var i = 0; i < input.length; i++) {
          var item = input[i];
          if((key == "" && item.indexOf(filterInput) >= 0) || (key != "" && item[key].indexOf(filterInput) >= 0)) {
            filteredResults.push(item);
            continue;
          }
        }
        input = filteredResults;
      }
      return input;
    }
  }
  
})();