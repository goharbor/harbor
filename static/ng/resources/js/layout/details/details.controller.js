(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .controller('DetailsController', DetailsController);
    
  DetailsController.$inject = ['$scope', '$location', '$routeParams'];
  
  function DetailsController($scope, $location, $routeParams) {
    var vm = this;
    vm.isOpen = false;
    vm.publicity = false;
    vm.isProjectMember = true;
    vm.closeRetrievePane = closeRetrievePane;   
    vm.togglePublicity = togglePublicity;
    
    function closeRetrievePane() {
      $scope.$broadcast('isOpen', false);
    }
    function togglePublicity(e) {
      vm.publicity = e.publicity;
      console.log('current project publicity:' + vm.publicity);
    }
  }
  
})();