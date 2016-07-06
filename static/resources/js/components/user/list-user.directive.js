(function() {
  
  'use strict';
  
  angular
    .module('harbor.user')
    .directive('listUser', listUser);
    
  ListUserController.$inject = ['$scope', 'ListUserService', 'DeleteUserService', '$filter', 'trFilter'];
  
  function ListUserController($scope, ListUserService, DeleteUserService, $filter, $trFilter) {

    $scope.subsSubPane = 226;
        
    var vm = this;
        
    vm.username = '';
    vm.searchUser = searchUser;
    vm.deleteUser = deleteUser;
    vm.confirmToDelete = confirmToDelete;
    vm.retrieve = retrieve;
    
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
    
    function link(scope, element, attrs) {
      
    }
  }
  
})();