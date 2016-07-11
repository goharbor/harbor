/*
    Copyright (c) 2016 VMware, Inc. All Rights Reserved.
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
        
        http://www.apache.org/licenses/LICENSE-2.0
        
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/
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
        };
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
