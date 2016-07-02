(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.admin.option')
    .controller('AdminOptionController', AdminOptionController);
  
  AdminOptionController.$inject = ['$scope', '$timeout', '$location'];
  
  function AdminOptionController($scope, $timeout, $location) {
    
    $scope.subsSubPane = 296;   
    
    var vm = this;
    vm.toggle = false;
    vm.target = 'users';
    vm.toggleAdminOption = toggleAdminOption;
        
    $scope.$on('$locationChangeSuccess', function(e) {
       if($location.path() === '') {
         vm.target = 'users';
         vm.toggle = false;
       }else{
         vm.target = 'system_management'; 
         vm.toggle = true;
       }
    });
        
    //Message dialog handler for admin-options.
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
    
    
    function toggleAdminOption(e) {
      switch(e.target) {
      case 'users':
        vm.toggle = false;
        break;
      case 'system_management':
        vm.toggle = true;
        break;
      }
      vm.target = e.target;
    }
  }
  
})();