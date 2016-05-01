(function() {

  'use strict';
  
  angular
    .module('harbor.project')
    .directive('listProject', listProject);
  
  ListProjectController.$inject = ['ListProjectService', 'CurrentUserService', '$scope'];
  
  function ListProjectController(ListProjectService, CurrentUserService, $scope) {
    var vm = this;
    
    vm.retrieve = retrieve;
    vm.reload = reload;
    
    vm.getCurrentUser = getCurrentUser;
    
    function reload() {
      $.when(vm.getCurrentUser())
        .done(function(e) {
          vm.retrieve(vm.projectName);
        });
    }
    
    vm.reload();
    
    $scope.$on('needToReload', function(e, val) {
      if(val) {
        vm.reload(vm.projectName);
      }
    });
    
    function retrieve(projectName) {
      ListProjectService({'is_public': 0, 'project_name': projectName})
        .success(listProjectSuccess)
        .error(listProjectFailed);
    }
    
    function listProjectSuccess(data, status) {
      vm.projects = data;
    }
    function listProjectFailed(e) {
      console.log('Failed in listProject:' + e);
    }
    
    function getCurrentUser() {
      CurrentUserService()
        .success(getCurrentUserSuccess)
        .error(getCurrentUserFailed);
    }
    
    function getCurrentUserSuccess(data, status) {
      vm.currentUser = data;
    }
    
    function getCurrentUserFailed(e) {
      console.log('Failed in getCurrentUser:' + e);
    }
    
  }
  
  function listProject() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/project/list-project.directive.html',
      'scope': {
        'projectName': '='
      },
      'controller': ListProjectController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
  
})();