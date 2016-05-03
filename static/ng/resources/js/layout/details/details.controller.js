(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .controller('DetailsController', DetailsController);
    
  DetailsController.$inject = ['ListProjectService', '$scope', '$location', '$routeParams'];
  
  function DetailsController(ListProjectService, $scope, $location, $routeParams) {
    var vm = this;
    vm.isOpen = false;
    vm.closeRetrievePane = closeRetrievePane;   
   
    function closeRetrievePane() {
      $scope.$broadcast('isOpen', false);
    }
  }
  
})();