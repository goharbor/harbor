(function() {
  
  'use strict';
  
  angular
    .module('harbor.replication')
    .directive('createPolicy', createPolicy);
  
  CreatePolicyController.$inject = ['$scope', 'ListReplicationPolicyService', 'ListDestinationService', 'UpdateDestinationService', 'PingDestinationService', 'CreateReplicationPolicyService', 'UpdateReplicationPolicyService', 'ListDestinationPolicyService','$location', 'getParameterByName', '$filter', 'trFilter'];
  
  function CreatePolicyController($scope, ListReplicationPolicyService, ListDestinationService, UpdateDestinationService, PingDestinationService, CreateReplicationPolicyService, UpdateReplicationPolicyService, ListDestinationPolicyService, $location, getParameterByName, $filter, trFilter) {
    var vm = this;
    
    //Since can not set value for textarea by using vm
    //use $scope for instead.
    $scope.replication = {};
    $scope.replication.policy = {};
    $scope.replication.destination = {};
    
    var vm0 = $scope.replication.policy;
    var vm1 = $scope.replication.destination;
            
    vm.selectDestination = selectDestination;
    vm.projectId = getParameterByName('project_id', $location.absUrl());
    
    $scope.$on('$locationChangeSuccess', function() {
      vm.projectId = getParameterByName('project_id', $location.absUrl());
    });
    
    vm.addNew = addNew;
    vm.edit = edit;
    vm.prepareDestination = prepareDestination;
    vm.create = create;
    vm.update = update;
    vm.pingDestination = pingDestination;
    
    vm.targetEditable = true;
        
    $scope.$watch('vm.destinations', function(current) {
      if(current) {
        vm1.selection = current[0]; 
        vm1.endpoint = vm1.selection.endpoint;
        vm1.username = vm1.selection.username;
        vm1.password = vm1.selection.password;
      }
    });
                                     
    function selectDestination(item) {
      vm1.selection = item;
      vm1.endpoint = item.endpoint;
      vm1.username = item.username;
      vm1.password = item.password;
    }
    
    function prepareDestination() {
      ListDestinationService('')
        .success(listDestinationSuccess)
        .error(listDestinationFailed);
    }

    function addNew() {       
      vm.targetEditable = true;
      $filter('tr')('add_new_policy', []);
      vm0.name = '';
      vm0.description = '';
      vm0.enabled = true;
    }
    
    function edit(policyId) {
      console.log('Edit policy ID:' + policyId);
      vm.policyId = policyId;
      vm.targetEditable = true;
      $filter('tr')('edit_policy', []);
      ListReplicationPolicyService(policyId)
        .success(listReplicationPolicySuccess)
        .error(listReplicationPolicyFailed);
    }
    
    function create(policy) {
      CreateReplicationPolicyService(policy)
        .success(createReplicationPolicySuccess)
        .error(createReplicationPolicyFailed);
    }
    
    function update(policy) {
      console.log('Update policy ID:' + vm.policyId);
      UpdateReplicationPolicyService(vm.policyId, policy)
        .success(updateReplicationPolicySuccess)
        .error(updateReplicationPolicyFailed);
        
      var targetId = vm1.selection.id;
      console.log('Update target ID:' + targetId);
      var target = {
        'name': vm1.selection.name,
        'endpoint': vm1.endpoint,
        'username': vm1.username,
        'password': vm1.password
      };
      UpdateDestinationService(targetId, target)
        .success(updateDestinationSuccess)
        .error(updateDestinationFailed);
    }
    
    function pingDestination() {
      var target = {
        'name': vm1.selection.name,
        'endpoint': vm1.endpoint,
        'username': vm1.username,
        'password': vm1.password
      };
      PingDestinationService(target)
        .success(pingDestinationSuccess)
        .error(pingDestinationFailed);
    }
    
    function listDestinationSuccess(data, status) {
      vm.destinations = data || [];
    }
    function listDestinationFailed(data, status) {
      console.log('Failed list destination:' + data);
    }
    
    function listDestinationPolicySuccess(data, status) {
      vm.targetEditable = true;
      for(var i in data) {
        if(data[i].enabled === 1) {
          vm.targetEditable = false;
          break;
        }
      }
      console.log('current target editable:' + vm.targetEditable + ', policy ID:' + vm.policyId);
    }
    
    function listDestinationPolicyFailed(data, status) {
      console.log('Failed list destination policy:' + data);
    }
    
    function listReplicationPolicySuccess(data, status) {
      console.log(data);
      var replicationPolicy = data;
      vm0.name = replicationPolicy.name;
      vm0.description = replicationPolicy.description;
      vm0.enabled = replicationPolicy.enabled == 1;
      vm.targetId = replicationPolicy.target_id;
     
      if(vm0.enabled) {
        vm.targetEditable = false;
      }else{
        ListDestinationPolicyService(vm.targetId)
         .success(listDestinationPolicySuccess)
         .error(listDestinationPolicyFailed);
      }
    }
    function listReplicationPolicyFailed(data, status) {
      console.log('Failed list replication policy:' + data);
    }
    function createReplicationPolicySuccess(data, status) {
      console.log('Successful create replication policy.');
      vm.reload();
    }
    function createReplicationPolicyFailed(data, status) {
      if(status === 409) {
        alert($filter('tr')('policy_already_exists', []));
      }
      console.log('Failed create replication policy.');
    }
    function updateReplicationPolicySuccess(data, status) {
      console.log('Successful update replication policy.');
      vm.reload();
    }
    function updateReplicationPolicyFailed(data, status) {
      console.log('Failed update replication policy.');
    }
    function updateDestinationSuccess(data, status) {
      console.log('Successful update destination.');
    }
    function updateDestinationFailed(data, status) {
      console.log('Failed update destination.');
    }
    function pingDestinationSuccess(data, status) {
      alert($filter('tr')('successful_ping_target', []));
    }
    function pingDestinationFailed(data, status) {
      alert($filter('tr')('failed_ping_target', []) + ':' + data);
    }
  }
  
  function createPolicy() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/replication/create-policy.directive.html',
      'scope': {
        'policyId': '@',
        'modalTitle': '@',
        'reload': '&',
        'action': '='
      },
      'link': link,
      'controller': CreatePolicyController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attr, ctrl) {  
      
      element.find('#createPolicyModal').on('show.bs.modal', function() {    
        
        scope.form.$setPristine();
        scope.form.$setUntouched();
        
        ctrl.prepareDestination();
        switch(ctrl.action) {
        case 'ADD_NEW':
          ctrl.addNew(); 
          break;
        case 'EDIT':
          ctrl.edit(ctrl.policyId); 
          break;
        }  
        scope.$apply();
      });  
                  
      ctrl.save = save;
    
      function save(form) {
        var postPayload = {
          'projectId': Number(ctrl.projectId),
          'targetId': form.destination.selection.id,
          'name': form.policy.name,
          'enabled': form.policy.enabled ? 1 : 0,
          'description': form.policy.description,
          'cron_str': '',
          'start_time': ''
        };
        switch(ctrl.action) {
        case 'ADD_NEW':
          ctrl.create(postPayload);
          break;
        case 'EDIT':
          ctrl.update(postPayload);
          break;
        }
        element.find('#createPolicyModal').modal('hide');
      }
    }
  }
  
})();