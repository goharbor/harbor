(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .controller('DetailsController', DetailsController);

  DetailsController.$inject = ['$scope'];

  function DetailsController($scope) {
    var vm = this;
          
    vm.publicity = false;
    vm.isProjectMember = false;
    
    vm.togglePublicity = togglePublicity;
    vm.target = 'repositories';
    
    function togglePublicity(e) {
      vm.publicity = e.publicity;
    }
  }
  
})();