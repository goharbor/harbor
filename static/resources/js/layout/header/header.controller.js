(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.header')
    .controller('HeaderController', HeaderController);
  
  HeaderController.$inject = ['$scope', '$window', 'getParameterByName', '$location', 'currentUser'];
  
  function HeaderController($scope, $window, getParameterByName, $location, currentUser) {
    var vm = this;
    vm.user = currentUser.get();
        
    if(location.pathname === '/dashboard') {
      vm.defaultUrl = '/dashboard';
    }else{
      vm.defaultUrl = '/';
    }
    
    $scope.$watch('vm.user', function(current) {
      if(current) {
        vm.defaultUrl = '/dashboard';
      }
    });
    
    if($window.location.search) {
      vm.searchInput = getParameterByName('q', $window.location.search);
      console.log('vm.searchInput at header:' + vm.searchInput);
    }
  }
  
})();