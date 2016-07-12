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
    .module('harbor.system.management')
    .directive('destination', destination);
    
  DestinationController.$inject = ['$scope', 'ListDestinationService', 'DeleteDestinationService', '$filter', 'trFilter'];
  
  function DestinationController($scope, ListDestinationService, DeleteDestinationService, $filter, trFilter) {
    
    $scope.subsSubPane = 276;
    $scope.subsTblBody = 40;
    var vm = this;
    
    vm.retrieve = retrieve;
    vm.search = search;
    vm.addDestination = addDestination;
    vm.editDestination = editDestination;
    vm.confirmToDelete = confirmToDelete;
    vm.deleteDestination = deleteDestination;
    
    vm.retrieve();
    
    function retrieve() {
      ListDestinationService('', vm.destinationName)
        .success(listDestinationSuccess)
        .error(listDestinationFailed);
    }
    
    function search() {
      vm.retrieve();
    }
    
    function addDestination() {
      vm.action = 'ADD_NEW';
      console.log('Action for destination:' + vm.action);
    }
    
    function editDestination(targetId) {
      vm.action = 'EDIT';
      vm.targetId = targetId;
      console.log('Action for destination:' + vm.action + ', target ID:' + vm.targetId);
    }
    
    function confirmToDelete(targetId, name) {
      vm.selectedTargetId = targetId;
      
      $scope.$emit('modalTitle', $filter('tr')('confirm_delete_destination_title'));
      $scope.$emit('modalMessage', $filter('tr')('confirm_delete_destination', [name]));
      
      var emitInfo = {
        'confirmOnly': false,
        'contentType': 'text/plain',
        'action': vm.deleteDestination
      };
      
      $scope.$emit('raiseInfo', emitInfo);
    }
    
    function deleteDestination() {
      DeleteDestinationService(vm.selectedTargetId)
        .success(deleteDestinationSuccess)
        .error(deleteDestinationFailed);
    }
    
    function listDestinationSuccess(data, status) {
      vm.destinations = data || [];
    }
    
    function listDestinationFailed(data, status) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_list_destination'));
      $scope.$emit('raiseError', true);
      console.log('Failed to list destination:' + data);
    }
    
    function deleteDestinationSuccess(data, status) {
      console.log('Successful delete destination.');
      vm.retrieve();
    }
    
    function deleteDestinationFailed(data, status) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_delete_destination') + data);
      $scope.$emit('raiseError', true);
      console.log('Failed to delete destination.');
    }   
  }
  
  function destination() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/system-management/destination.directive.html',
      'scope': true,
      'link': link,
      'controller': DestinationController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      element.find('#txtSearchInput').on('keydown', function(e) {
        if($(this).is(':focus') && e.keyCode === 13) {
          ctrl.retrieve();
        }
      });
    }
  }
  
})();