(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.destination')
    .factory('PingDestinationService', PingDestinationService);
    
  PingDestinationService.$inject = ['$http'];
  
  function PingDestinationService($http) {
    return pingDestination;
    function pingDestination(target) {
      var payload = {};
      if(target['id']) {
        payload = {'id': target['id']};
      }else {
        payload = {
          'name': target['name'],
          'endpoint': target['endpoint'],
          'username': target['username'],
          'password': target['password']
        };
      }
      
      return $http({
           'method': 'POST',
              'url': '/api/targets/ping',
          'headers': {'Content-Type': 'application/x-www-form-urlencoded'},
          'transformRequest': function(obj) {
              var str = [];
              for(var p in obj) {
                str.push(encodeURIComponent(p) + "=" + encodeURIComponent(obj[p]));
              }
              return str.join("&");
          },
          'timeout': 30000,
          'data': payload
        });
    }
  }
  
})();