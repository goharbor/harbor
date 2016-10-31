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
    
    $scope.$on('modalTitle', function(e, val) {
      vm.modalTitle = val;
    });
    
    $scope.$on('modalMessage', function(e, val) {
      vm.modalMessage = val;
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
  }
  
})();