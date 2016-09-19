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
    .module('harbor.details')
    .controller('DetailsController', DetailsController);

  DetailsController.$inject = ['$scope', '$timeout', '$window'];

  function DetailsController($scope, $timeout, $window) {
    var vm = this;
          
    vm.isPublic = 0;
    vm.isProjectMember = false;
    
    vm.togglePublicity = togglePublicity;
    
    vm.sectionHeight = {'min-height': '579px'};
    
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
        };
        vm.contentType = val.contentType;
        vm.confirmOnly = val.confirmOnly;
       
        $scope.$broadcast('showDialog', true);
      }
    });
    
    $scope.$on('projectChanged', function(e, val) {
      if(val) {
        $scope.$broadcast('retrieveData', true);
      }
    });
     
    function togglePublicity(e) {      
      vm.isPublic = e.isPublic;
      $window.location='/project?is_public=' + vm.isPublic;
      return;
    }
  }
  
})();
