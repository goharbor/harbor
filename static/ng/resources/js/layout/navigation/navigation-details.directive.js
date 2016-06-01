(function() {

  'use strict';

  angular
    .module('harbor.layout.navigation')
    .directive('navigationDetails', navigationDetails);
  
  NavigationDetailsController.$inject = ['$window', '$location', '$scope', 'getParameterByName', 'CurrentProjectMemberService'];
  
  function NavigationDetailsController($window, $location, $scope, getParameterByName, CurrentProjectMemberService) {
    var vm = this;    
     
    vm.projectId = getParameterByName('project_id', $location.absUrl());

    $scope.$on('$locationChangeSuccess', function() {
      vm.projectId = getParameterByName('project_id', $location.absUrl());
      CurrentProjectMemberService(vm.projectId)
        .success(getCurrentProjectMemberSuccess)
        .error(getCurrentProjectMemberFailed);
    });
    
    function getCurrentProjectMemberSuccess(data, status) {
      console.log('Successful get current project member:' + status);
      vm.isProjectMember = true;
    }
    
    function getCurrentProjectMemberFailed(data, status) {
      console.log('Failed get current project member:' + status);
      vm.isProjectMember = false;
    }
    
    vm.path = $location.path();
  }
  
  function navigationDetails() {
    var directive = {
      restrict: 'E',
      templateUrl: '/ng/navigation_detail',
      link: link,
      scope: {
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