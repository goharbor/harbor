(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.search')
    .controller('SearchController', SearchController);
   
  SearchController.$inject = ['$location', 'SearchService', '$scope', '$filter', 'trFilter', 'getParameterByName'];
  
  function SearchController($location, SearchService, $scope, $filter, trFilter, getParameterByName) {
    var vm = this;
    
    vm.q = getParameterByName('q', $location.absUrl());
    console.log('vm.q:' + vm.q);
    SearchService(vm.q)
      .success(searchSuccess)
      .error(searchFailed);
  
    //Error message dialog handler for search.
    $scope.$on('modalTitle', function(e, val) {
      vm.modalTitle = val;
    });
    
    $scope.$on('modalMessage', function(e, val) {
      vm.modalMessage = val;
    });
       
    $scope.$on('raiseError', function(e, val) {
      if(val) {   
        vm.action = function() {
          $scope.$broadcast('showDialog', false);
        };
        vm.contentType = 'text/plain';
        vm.confirmOnly = true;      
        $scope.$broadcast('showDialog', true);
      }
    });
    
    function searchSuccess(data, status) {
      vm.repository = data['repository'];
      vm.project = data['project'];
    }
    
    function searchFailed(data, status) {
      
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_in_search'));
      $scope.$emit('raiseError', true);
      
      console.log('Failed search:' + data);
    }
  }
  
})();