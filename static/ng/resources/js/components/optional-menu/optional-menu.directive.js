(function() {
  
  'use strict';
  
  angular
    .module('harbor.optional.menu')
    .directive('optionalMenu', optionalMenu);

  OptionalMenuController.$inject = ['$scope'];

  function OptionalMenuController($scope, $timeout) {
    var vm = this;
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