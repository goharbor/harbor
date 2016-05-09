(function() {
  
  'use strict';
  
  angular
    .module('harbor.optional.menu')
    .directive('optionalMenu', optionalMenu);
  
  OptionalMenuController.$inject = ['$scope'];
  
  function OptionalMenuController($scope, CurrentUserService) {
    var vm = this;
    vm.username = 'abcde';
    $scope.$watch('vm.username', function(current) {
      if(current) {
        vm.username = current;
        console.log('vm.username:' + current);
      }
    });
  }
  
  function optionalMenu() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/optional-menu/optional-menu.directive.html',
      'scope': {
        'isLoggedIn': '=',
        'username': '='
      },
      'link': link,
      'controller': OptionalMenuController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      scope.$watch('vm.isLoggedIn', function(current) {
        if(current) {
          ctrl.isLoggedIn = current;
          console.log('vm.isLoggedIn:' + current);
        }
      });
      
      
    }
  }
  
})();