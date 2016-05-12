(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.log')
    .factory('ListIntegratedLogService', ListIntegratedLogService);
    
  ListIntegratedLogService.$inject = ['$http', '$q', '$timeout'];
    
  function ListIntegratedLogService($http, $q, $timeout) {
    
    var mockData = [
      {
        'task_name': 'create',
        'details': 'created myproject/ubuntu',
        'user': 'kunw',
        'creation_time': '2016-05-10 17:53:25'
      },
      {
        'task_name': 'push',
        'details': 'pushed myproject/mysql',
        'user': 'kunw',
        'creation_time': '2016-05-10 16:25:15'
      },
      {
        'task_name': 'pull',
        'details': 'pulled myrepo/nginx',
        'user': 'user1',
        'creation_time': '2016-05-11 10:42:43'
      },
      {
        'task_name': 'delete',
        'details': 'deleted myrepo/golang',
        'user': 'user1',
        'creation_time': '2016-05-11 12:21:35'
      }
    ];
   
    function async() {
      
      var deferred = $q.defer();   
      
      $timeout(function() {
        deferred.resolve(mockData);
      }, 500);
      
      return deferred.promise;
    }
    
    return listIntegratedLog;
    
    function listIntegratedLog() {
      return async();
    }
    
  }
  
  
  
  
})();