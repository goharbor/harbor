(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.project.member')
    .controller('ProjectMemberController', ProjectMemberController);
    
  ProjectMemberController.$inject = ['$scope'];    
    
  function ProjectMemberController($scope) {
     var vm = this;
     $scope.$on('currentProjectId', function(e, val) {
       console.log('received currentProjectId: ' + val + ' in ProjectMemberController');
       vm.projectId = val;
     });
  }
  
})();