(function() {

  'use strict';

  angular
    .module('harbor.layout.navigation')
    .directive('navigationDetails', navigationDetails);
  
  NavigationDetailsController.$inject = ['$window', '$location', '$scope', '$route'];
  
  function NavigationDetailsController($window, $location, $scope, $route) {
    var vm = this;    
    
    $scope.$watch('vm.selectedProject', function(current, origin) {
      if(current) {
        vm.projectId = current.ProjectId;
      }
    });
    
    vm.url = $location.url();
  }
  
  function navigationDetails() {
    var directive = {
      restrict: 'E',
      templateUrl: '/static/ng/resources/js/layout/navigation/navigation-details.directive.html',
      link: link,
      scope: {
        'selectedProject': '='
      },
      replace: true,
      controller: NavigationDetailsController,
      controllerAs: 'vm',
      bindToController: true
    };
    
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      
      var visited = ctrl.url.substring(1);
      
      if(visited.indexOf('?') >= 0) {
        visited = ctrl.url.substring(1, ctrl.url.indexOf('?') - 1);
      }
      
      scope.$watch('vm.selectedProject', function(current) {
        if(current) {
          element.find('a').removeClass('active');
          element.find('a:first').addClass('active');
        }
      });
     
      element.find('a[tag*="' + visited + '"]').addClass('active');
      element.find('a').on('click', click);
      
      function click(event) {
        element.find('a').removeClass('active');
        $(event.target).addClass('active');
      }
     
    }
   
  }
  
})();