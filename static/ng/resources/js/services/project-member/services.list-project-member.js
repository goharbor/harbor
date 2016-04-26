(function() {
  
  'use strict';
  
   angular
    .module('harbor.services.project.member')
    .service('ListProjectMemberService', ListProjectMemberService);
   
  ListProjectMemberService.$inject = ['$http', '$log'];
 
  function ListProjectMemberService($http, $log) {
    
    return ListProjectMember;
    
    function ListProjectMember(projectId, queryParams) {
      console.log('project_member project_id:' + projectId);
      var username = queryParams.username;
      return $http
        .get('/api/projects/' + projectId + '/members', {
          params: {
            'username': username
          }
        });
    }
  }
  
})();