(function() {
  
  'use strict';
  
  angular
    .module('harbor.replication')
    .directive('createPolicy', createPolicy);
  
  CreatePolicyController.$inject = ['$scope', 'ListDestinationService', 'CreateReplicationPolicyService', '$location', 'getParameterByName'];
  
  function CreatePolicyController($scope, ListDestinationService, CreateReplicationPolicyService, $location, getParameterByName) {
    var vm = this;
    
    //Since can not set value for textarea by using vm
    //use $scope for instead.
    $scope.replication = {};
    $scope.replication.policy = {};
    
    var vm0 = $scope.replication;
    var vm1 = $scope.replication.policy;
    vm.projectId = getParameterByName('project_id', $location.absUrl());
    vm.prepare = prepare;
    vm.prepare();
    vm.createPolicy = createPolicy;
   
    $scope.$watch('vm.destinations', function(current) {
      if(current) {
        console.log('destination:' + angular.toJson(current));
        vm0.destination = current[0]; 
      }
    });
    
    function prepare() {
      vm1.name = 'name';
      vm1.description = 'test';
      vm1.enabled = true;
     
      ListDestinationService()
        .success(listDestinationSuccess)
        .error(listDestinationFailed);
    }
    
    function createPolicy(policy) {
      CreateReplicationPolicyService(policy)
        .success(createReplicationPolicySuccess)
        .error(createReplicationPolicyFailed);
    }
    
    function listDestinationSuccess(data, status) {
      vm.destinations = data;
    }
    function listDestinationFailed(data, status) {
      console.log('Failed list destination:' + data);
    }
    function createReplicationPolicySuccess(data, status) {
      console.log('Successful create replication policy.');
      vm.clearUp();
    }
    function createReplicationPolicyFailed(data, status) {
      console.log('Failed create replication policy.');
    }
  }
  
  function createPolicy() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/replication/create-policy.directive.html',
      'scope': {
        'reload': '&'
      },
      'link': link,
      'controller': CreatePolicyController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attr, ctrl) {
      ctrl.save = save;
      ctrl.clearUp = clearUp;
      function save(form) {
        console.log(angular.toJson(form));
        var postPayload = {
          'projectId': Number(ctrl.projectId),
          'targetId': form.destination.id,
          'name': form.policy.name,
          'enabled': form.policy.enabled ? 1 : 0,
          'description': form.policy.description,
          'cron_str': '',
          'start_time': ''
        };
        ctrl.createPolicy(postPayload, clearUp);
      }
      
      function clearUp() {
        element.find('#createPolicyModal').modal('hide');
        ctrl.reload();
      }
    }
  }
  
})();