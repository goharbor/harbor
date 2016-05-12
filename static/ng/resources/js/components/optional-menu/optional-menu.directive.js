(function() {
  
  'use strict';
  
  angular
    .module('harbor.optional.menu')
    .directive('optionalMenu', optionalMenu);

  OptionalMenuController.$inject = ['$scope', '$window', 'LogOutService'];

  function OptionalMenuController($scope, $window, LogOutService) {
    var vm = this;
    vm.logOut = logOut;
    function logOut() {
      LogOutService()
        .success(logOutSuccess)
        .error(logOutFailed);
    }
    function logOutSuccess(data, status) {
      $window.location.href= '/ng';
    }
    function logOutFailed(data, status) {
      console.log('Failed to log out:' + data);
    }
  }
  
  function optionalMenu() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/optional-menu/optional-menu.directive.html',
      'link': link,
      'scope': true,
      'controller': OptionalMenuController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    function link(scope, element, attrs, ctrl) {
      ctrl.isLoggedIn = false;
      scope.$on('currentUser', function(e, val) {
        if(val != null) {
          ctrl.isLoggedIn = true;
          ctrl.username = val.username;
        }
        scope.$apply();
      });
    }
  }
  
})();