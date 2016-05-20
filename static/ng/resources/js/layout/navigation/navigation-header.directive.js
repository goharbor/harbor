(function() {

  'use strict';

  angular
    .module('harbor.layout.navigation')
    .directive('navigationHeader', navigationHeader);
  
  NavigationHeaderController.$inject = ['$window', '$scope', 'currentUser', '$timeout'];
    
  function NavigationHeaderController($window, $scope, currentUser, $timeout) {
    var vm = this;
    
    vm.isShow = false;
    vm.isAdmin = false;
    
    $timeout(function() {
      vm.user = currentUser.get();
      if(angular.isDefined(vm.user)) {
          vm.isShow = true;
          if(vm.user.HasAdminRole === 1) {
            vm.isAdmin = true;
          }
      }
    });
    vm.url = $window.location.pathname;    
  }
  
  function navigationHeader() {
    var directive = {
      restrict: 'E',
      templateUrl: '/static/ng/resources/js/layout/navigation/navigation-header.directive.html',
      link: link,
      scope: true,
      controller: NavigationHeaderController,
      controllerAs: 'vm',
      bindToController: true
    }
    
    return directive;
   
    function link(scope, element, attrs, ctrl) {     
      element.find('a').removeClass('active');
      var visited = ctrl.url;
      if (visited != "/ng") {
         element.find('a[href*="' + visited + '"]').addClass('active'); 
      }      
      element.on('click', click);
      function click(event) {
        element.find('a').removeClass('active');
        $(event.target).not('span').addClass('active');
      }     
    }
   
  }
  
})();