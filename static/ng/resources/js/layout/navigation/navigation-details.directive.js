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
    
    vm.path = $location.path();
  }
  
  function navigationDetails() {
    var directive = {
      restrict: 'E',
      templateUrl: '/ng/navigation_detail',
      link: link,
      scope: {
        'selectedProject': '=',
        'target': '='
      },
      replace: true,
      controller: NavigationDetailsController,
      controllerAs: 'vm',
      bindToController: true
    };
    
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      
      var visited = ctrl.path.substring(1);  
      if(visited.indexOf('?') >= 0) {
        visited = ctrl.url.substring(1, ctrl.url.indexOf('?'));
      }
      
      if(visited) {
        element.find('a[tag="' + visited + '"]').addClass('active');
      }else{
        element.find('a:first').addClass('active');
      }

      element.find('a').on('click', click);
            
      function click(event) {
        element.find('a').removeClass('active');
        $(event.target).addClass('active');
        ctrl.target = $(this).attr('tag');
        scope.$apply();
      }
     
    }
   
  }
  
})();