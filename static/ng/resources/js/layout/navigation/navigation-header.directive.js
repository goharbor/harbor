(function() {

  'use strict';

  angular
    .module('harbor.layout.navigation')
    .directive('navigationHeader', navigationHeader);
  
  NavigationHeaderController.$inject = ['$window', '$scope'];
    
  function NavigationHeaderController($window, $scope) {
    var vm = this;
    vm.url = $window.location.pathname;   
    vm.isAdmin = false;
    $scope.$on('currentUser', function(e, val) {
      if(val.HasAdminRole === 1) {
        vm.isAdmin = true;
      }
      $scope.$apply();
    });
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