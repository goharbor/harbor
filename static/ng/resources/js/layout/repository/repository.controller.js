(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.repository')
    .controller('RepositoryController', RepositoryController);
   
  RepositoryController.$inject = ['$scope'];
  
  function RepositoryController($scope) {
    var vm = this;
    
    $scope.$on('currentProjectId', function(e, val){
      console.log('received currentProjecjtId: ' + val + ' in RepositoryController');  
      vm.projectId = val;
    }); 
   
  }
  
})();