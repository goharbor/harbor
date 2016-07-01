(function() {
    
  'use strict';
  
  angular
    .module('harbor.top.repository')
    .directive('topRepository', topRepository);
    
  TopRepositoryController.$inject = ['$scope', 'ListTopRepositoryService', '$filter', 'trFilter'];
  
  function TopRepositoryController($scope, ListTopRepositoryService, $filter, trFilter) {
    var vm = this;
    
    ListTopRepositoryService(5)
      .success(listTopRepositorySuccess)
      .error(listTopRepositoryFailed);
      
    function listTopRepositorySuccess(data) {
      vm.top10Repositories = data || [];
    }

    function listTopRepositoryFailed(data, status) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_get_top_repo'));
      $scope.$emit('raiseError', true);
      console.log('Failed get top repo:' + data);
    }
  }
  
  function topRepository() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/top-repository/top-repository.directive.html',
      'controller': TopRepositoryController,
      'scope' : {
        'customBodyHeight': '='
      },
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
    
})();
