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
    vm.resultCount = 0;
    vm.publicity = 0;
    
    function reload() {
      $.when(vm.getCurrentUser())
        .done(function(e) {
          vm.retrieve();
        });
    }
    
    vm.reload();
    
    $scope.$on('needToReload', function(e, val) {
      if(val) {
        vm.reload();
      }
    });
    
    function retrieve() {
      ListProjectService(vm.projectName, vm.publicity)
        .success(listProjectSuccess)
        .error(listProjectFailed);
    }
    
    function listProjectSuccess(data, status) {
      vm.projects = data;
      if(data) {
        vm.resultCount = vm.projects.length;  
      }
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
        'projectName': '=',
        'publicity': '=',
        'resultCount': '='
      },
      'controller': ListProjectController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
  
})();