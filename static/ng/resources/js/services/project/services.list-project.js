(function() {
  'use strict';
 
   angular
    .module('harbor.services.project')
    .factory('ListProjectService', ListProjectService);
  
  ListProjectService.$inject = ['$http', '$log'];
  
  function ListProjectService($http, $log) {
    
    return ListProject;
    
    function ListProject(queryParams) {
      
      $log.info(queryParams);
      
      var isPublic = queryParams.isPublic;      
      var projectName = queryParams.projectName;
      return $http({
          method: 'GET',
          url: '/api/projects',
          headers: {'Content-Type': 'application/x-www-form-urlencoded'},
          transformRequest: function(obj) {
              var str = [];
              for(var p in obj)
              str.push(encodeURIComponent(p) + "=" + encodeURIComponent(obj[p]));
              return str.join("&");
          },
          data: {'is_public': isPublic, 'project_name': projectName}
        });      
    }
  }
})();