(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .controller('DetailsController', DetailsController);

  DetailsController.$inject = ['$scope', '$timeout'];

  function DetailsController($scope, $timeout) {
    var vm = this;
          
    vm.publicity = false;
    vm.isProjectMember = false;
    
    vm.togglePublicity = togglePublicity;
    vm.target = 'repositories';
    
    vm.sectionDefaultHeight = {'min-height': '579px'};
    
    //Message dialog handler for details.
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
        
        $timeout(function() {
          $scope.$broadcast('showDialog', true);  
        }, 350);
      }
    });  
    
    $scope.$on('raiseInfo', function(e, val) {
      if(val) {
        vm.action = function() {
          val.action();
          $scope.$broadcast('showDialog', false);
        }
        vm.contentType = val.contentType;
        vm.confirmOnly = val.confirmOnly;
       
        $scope.$broadcast('showDialog', true);
      }
    });
    
    function togglePublicity(e) {
      vm.publicity = e.publicity;
    }
  }
  
})();