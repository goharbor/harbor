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
    .module('harbor.user')
    .directive('listUser', listUser);
    
  ListUserController.$inject = ['$scope', 'ListUserService', 'DeleteUserService', 'currentUser', '$filter', 'trFilter'];
  
  function ListUserController($scope, ListUserService, DeleteUserService, currentUser, $filter, $trFilter) {

    $scope.subsSubPane = 226;
        
    var vm = this;
        
    vm.username = '';
    vm.searchUser = searchUser;
    vm.deleteUser = deleteUser;
    vm.confirmToDelete = confirmToDelete;
    vm.retrieve = retrieve;

    vm.currentUser = currentUser.get();
    
    vm.retrieve();
    
    function searchUser() {
      vm.retrieve();
    }
    
    function deleteUser() {
      DeleteUserService(vm.selectedUserId)
        .success(deleteUserSuccess)
        .error(deleteUserFailed);
    }
    
    function confirmToDelete(userId, username) {
      vm.selectedUserId = userId;
     
      $scope.$emit('modalTitle', $filter('tr')('confirm_delete_user_title'));
      $scope.$emit('modalMessage', $filter('tr')('confirm_delete_user', [username]));
      
      var emitInfo = {
        'confirmOnly': false,
        'contentType': 'text/plain',
        'action': vm.deleteUser
      };
      
      $scope.$emit('raiseInfo', emitInfo);
    }
    
    function retrieve() {
      ListUserService(vm.username)
        .success(listUserSuccess)
        .error(listUserFailed);
    }
    
    function deleteUserSuccess(data, status) {
      console.log('Successful delete user.');
      vm.retrieve();
    }
    
    function deleteUserFailed(data, status) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_delete_user'));
      $scope.$emit('raiseError', true);
      console.log('Failed to delete user.');
    }
    
    function listUserSuccess(data, status) {
      vm.users = data;
    }
    
    function listUserFailed(data, status) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_list_user'));
      $scope.$emit('raiseError', true);
      console.log('Failed to list user:' + data);
    }      
  }
  
  function listUser() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/user/list-user.directive.html',
      'link': link,
      'controller': ListUserController,
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