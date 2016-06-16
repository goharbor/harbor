(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.project.member')
    .factory('CurrentProjectMemberService', CurrentProjectMemberService);
  
  CurrentProjectMemberService.$inject = ['$http', '$log'];  
    
  function CurrentProjectMemberService($http, $log) {
    return currentProjectMember;
    
    function currentProjectMember(projectId) {
      return $http
        .get('/api/projects/' + projectId + '/members/current');
    }
  }
  
})();