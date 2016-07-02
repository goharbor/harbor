(function() {
  
  'use strict';
  
  angular
    .module('harbor.replication')
    .directive('createPolicy', createPolicy);
  
  CreatePolicyController.$inject = ['$scope', 'ListReplicationPolicyService', 'ListDestinationService', 'CreateDestinationService', 'UpdateDestinationService', 'PingDestinationService', 'CreateReplicationPolicyService', 'UpdateReplicationPolicyService', 'ListDestinationPolicyService','$location', 'getParameterByName', '$filter', 'trFilter', '$q', '$timeout'];
  
  function CreatePolicyController($scope, ListReplicationPolicyService, ListDestinationService, CreateDestinationService, UpdateDestinationService, PingDestinationService, CreateReplicationPolicyService, UpdateReplicationPolicyService, ListDestinationPolicyService, $location, getParameterByName, $filter, trFilter, $q, $timeout) {
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
    vm.checkDestinationPolicyStatus = checkDestinationPolicyStatus;
  
    vm.targetEditable = true;
    vm.checkedAddTarget = false;
    
    vm.notAvailable = false;
    vm.pingAvailable = true;
    vm.pingMessage = '';
    
    vm.toggleInProgress0 = false;
    vm.toggleInProgress1 = false;
    
    vm.closeError = closeError;
    vm.toggleErrorMessage = false;
    vm.errorMessages = [];
              
    $scope.$watch('vm.destinations', function(current) {
      if(current) {
        if(!angular.isArray(current) || current.length === 0) {
          vm.notAvailable = true;
          return;
        }
        if(!angular.isDefined(vm1.selection)) {
          vm1.selection = current[0];
          vm1.endpoint = current[0].endpoint;
          vm1.username = current[0].username;
          vm1.password = current[0].password;
        }
      }
    });
    
    $scope.$watch('vm.checkedAddTarget', function(current) {
      if(current) {
        vm.targetEditable = true;
        vm1.name = '';
        vm1.endpoint = '';
        vm1.username = '';
        vm1.password = '';
        vm.pingMessage = '';
      }        
    });
        
    $scope.$watch('vm.targetId', function(current) {
      if(current) {          
        vm1.selection.id = current;
      }
    });    
        
    $scope.$watch('replication.destination.endpoint', function(current) {
      if(current) {
        vm.notAvailable = false;
      }else{
        vm.notAvailable = true; 
      }
    });
                                             
    function selectDestination(item) {
      vm1.selection = item;
      if(angular.isDefined(item)) {
        vm.targetId = item.id;
        vm1.endpoint = item.endpoint;
        vm1.username = item.username;
        vm1.password = item.password;
      }
    }
    
    function prepareDestination() {
      ListDestinationService('')
        .success(listDestinationSuccess)
        .error(listDestinationFailed);
    }

    function addNew() {       
      vm.modalTitle = $filter('tr')('add_new_policy', []);
     
      vm0.name = '';
      vm0.description = '';
      vm0.enabled = true;
    }
    
    function edit(policyId) {
      console.log('Edit policy ID:' + policyId);
      vm.policyId = policyId;
      
      vm.modalTitle = $filter('tr')('edit_policy', []);

      ListReplicationPolicyService(policyId)
        .success(listReplicationPolicySuccess)
        .error(listReplicationPolicyFailed);      
    }
    
    function create(policy) {
      vm.policy = policy;
      saveDestination();
    }
    
    function saveDestination() {
      
      var target = {
        'name'    : vm1.name,
        'endpoint': vm1.endpoint,
        'username': vm1.username,
        'password': vm1.password
      };
     
      if(vm.checkedAddTarget){
        CreateDestinationService(target.name, target.endpoint, target.username, target.password)
          .success(createDestinationSuccess)
          .error(createDestinationFailed);
      }else{
        vm.policy.targetId = vm1.selection.id || vm.destinations[0].id;
        saveOrUpdatePolicy();
      }
    }
    
    function saveOrUpdatePolicy() {
      vm.toggleInProgress1 = true;
      
      switch(vm.action) {
      case 'ADD_NEW':
        CreateReplicationPolicyService(vm.policy)
          .success(createReplicationPolicySuccess)
          .error(createReplicationPolicyFailed);
        break;
      case 'EDIT':
        UpdateReplicationPolicyService(vm.policyId, vm.policy)
          .success(updateReplicationPolicySuccess)
          .error(updateReplicationPolicyFailed);
        break;
      default:
        vm.toggleInProgress = false;
      }
    }
    
    function update(policy) {
      vm.policy = policy;        
      if(vm.targetEditable) {
        vm.policy.targetId = vm1.selection.id;
        saveOrUpdatePolicy();
      }
    }
        
    function pingDestination() {
            
      var target = {
        'endpoint': vm1.endpoint,
        'username': vm1.username,
        'password': vm1.password
      };
      
      if(vm.checkedAddTarget) {
        target.name = vm1.name;
      }
      
      vm.pingAvailable = false;
      vm.pingMessage = $filter('tr')('pinging_target');
      vm.toggleInProgress0 = true;
      
      PingDestinationService(target)
        .success(pingDestinationSuccess)
        .error(pingDestinationFailed);
    }
    
    function checkDestinationPolicyStatus() {
      ListDestinationPolicyService(vm.targetId)
        .success(listDestinationPolicySuccess)
        .error(listDestinationPolicyFailed);
    }
    
    function closeError() {
      vm.toggleErrorMessage = false;
    }
        
    function listDestinationSuccess(data, status) {
      vm.destinations = data || [];     
    }
    function listDestinationFailed(data, status) {
      vm.errorMessages.push($filter('tr')('failed_get_destination'));
      console.log('Failed get destination:' + data);
    }
    
    function listDestinationPolicySuccess(data, status) {
      if(vm.action === 'EDIT') {
        vm.targetEditable = true;
        for(var i in data) {
          if(data[i].enabled === 1) {
            vm.targetEditable = false;
            break;
          }
        }
      }
      console.log('current target editable:' + vm.targetEditable + ', policy ID:' + vm.policyId);
    }
    
    function listDestinationPolicyFailed(data, status) {
      vm.errorMessages.push($filter('tr')('failed_get_destination_policies'));
      console.log('Failed list destination policy:' + data);
    }
    
    function listReplicationPolicySuccess(data, status) {

      var replicationPolicy = data;
      
      vm.targetId = replicationPolicy.target_id;
      
      vm0.name = replicationPolicy.name;
      vm0.description = replicationPolicy.description;
      vm0.enabled = (replicationPolicy.enabled == 1);
      
      angular.forEach(vm.destinations, function(item) {
        if(item.id === vm.targetId) {
          vm1.endpoint = item.endpoint;
          vm1.username = item.username;
          vm1.password = item.password;
        }
      });
      
      vm.checkDestinationPolicyStatus();
    }
    function listReplicationPolicyFailed(data, status) {
      vm.errorMessages.push($filter('tr')('failed_get_replication_policy') + data);
      console.log('Failed list replication policy:' + data);
    }
    function createReplicationPolicySuccess(data, status) {
      vm.toggleInProgress1 = false;
      console.log('Successful create replication policy.');
      vm.reload();
    }
    function createReplicationPolicyFailed(data, status) {
      vm.toggleInProgress1 = false;
      if(status === 409) {
        vm.errorMessages.push($filter('tr')('policy_already_exists'));
      }else{
        vm.errorMessages.push($filter('tr')('failed_create_replication_policy') + data);
      }     
      console.log('Failed create replication policy.');
    }
    function updateReplicationPolicySuccess(data, status) {
      vm.toggleInProgress1 = false;
      console.log('Successful update replication policy.');
      vm.reload();
    }
    function updateReplicationPolicyFailed(data, status) {
      vm.toggleInProgress1 = false;
      if(status === 409) {
        vm.errorMessages.push($filter('tr')('policy_already_exists')); 
      }else{
        vm.errorMessages.push($filter('tr')('failed_update_replication_policy') + data);
      }
      console.log('Failed update replication policy.');
    }
    function createDestinationSuccess(data, status, headers) {
      var content = headers('Location');
      vm.policy.targetId = Number(content.substr(content.lastIndexOf('/') + 1));
      console.log('Successful create destination, targetId:' + vm.policy.targetId);
      saveOrUpdatePolicy();
    }
    function createDestinationFailed(data, status) {
      vm.errorMessages.push($filter('tr')('failed_create_destination') + data);
      console.log('Failed create destination.');
    }
    function updateDestinationSuccess(data, status) {
      console.log('Successful update destination.');
      vm.policy.targetId = vm1.selection.id;
      saveOrUpdatePolicy();
    }
    function updateDestinationFailed(data, status) {
      vm.errorMessages.push($filter('tr')('failed_update_destination') + data);
      $scope.$broadcast('showDialog', true);
      console.log('Failed update destination.');
    }
    function pingDestinationSuccess(data, status) {
      vm.pingAvailable = true;
      vm.pingMessage = $filter('tr')('successful_ping_target', []);
      vm.toggleInProgress0 = false;
    }
    function pingDestinationFailed(data, status) {
      vm.pingAvailable = true;
      vm.pingMessage = $filter('tr')('failed_ping_target', []) + (data && data.length > 0 ? ':' + data : '.');
      vm.toggleInProgress0 = false;
    }
  }
  
  function createPolicy($timeout) {
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
        scope.$apply(function() {
          scope.form.$setPristine();
          scope.form.$setUntouched();
          
          scope.$watch('vm.checkedAddTarget', function(current, origin) {
            if(origin) {
              var d = scope.replication.destination;
              if(angular.isDefined(d) && angular.isDefined(d.selection)) {
                d.endpoint = d.selection.endpoint;
                d.username = d.selection.username;
                d.password = d.selection.password;
                ctrl.checkDestinationPolicyStatus();
              }
            }
          }); 
          
          scope.$watch('vm.errorMessages', function(current) {
            if(current && current.length > 0) {
              ctrl.toggleErrorMessage = true;
            }
          }, true);
            
          ctrl.checkedAddTarget = false;
          ctrl.targetEditable = true;
          
          ctrl.notAvailable = false;
          
          ctrl.pingMessage = '';
          ctrl.pingAvailable = true;
          
          ctrl.toggleErrorMessage = false;
          ctrl.errorMessages = [];
          
          ctrl.prepareDestination();
          
          switch(ctrl.action) {
          case 'ADD_NEW':
            ctrl.addNew(); 
            break;
          case 'EDIT':
            ctrl.edit(ctrl.policyId); 
            break;
          }   
        });
      });               
                  
      ctrl.save = save;
    
      function save(form) {
        
        ctrl.toggleErrorMessage = false;
        ctrl.errorMessages = [];
        
        var postPayload = {
          'projectId': Number(ctrl.projectId),
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
        $timeout(function() {      
          if(!ctrl.toggleErrorMessage) {
            element.find('#createPolicyModal').modal('hide');
          }
        }, 50);
      }
    }
  }
  
})();