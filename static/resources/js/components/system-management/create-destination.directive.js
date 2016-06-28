(function() {
  
  'use strict';
  
  angular
    .module('harbor.system.management')
    .directive('createDestination', createDestination);
    
  CreateDestinationController.$inject = ['$scope', 'ListDestinationService', 'CreateDestinationService', 'UpdateDestinationService', 'PingDestinationService', 'ListDestinationPolicyService', '$filter', 'trFilter'];
  
  function CreateDestinationController($scope, ListDestinationService, CreateDestinationService, UpdateDestinationService, PingDestinationService, ListDestinationPolicyService, $filter, trFilter) {
    var vm = this;
    
    $scope.destination = {};
    
    var vm0 = $scope.destination;
    vm.addNew = addNew;
    vm.edit = edit;
    vm.create = create;
    vm.update = update;
    vm.pingDestination = pingDestination;
    
    vm.editable = true;
    vm.notAvailable = true;
    vm.pingAvailable = true;
    vm.pingMessage = '';
        
    $scope.$watch('destination.endpoint', function(current) {
      if(current) {
        vm.notAvailable = false;
      }else{
        vm.notAvailable = true;
      }
    });
        
    function addNew() {
      vm.editable = true;
      vm.modalTitle = $filter('tr')('add_new_destination', []);
      vm0.name = '';
      vm0.endpoint = '';
      vm0.username = '';
      vm0.password = '';
    }
    
    function edit(targetId) {
      vm.editable = true;
      vm.modalTitle = $filter('tr')('edit_destination', []);
      ListDestinationService(targetId)
        .success(getDestinationSuccess)
        .error(getDestinationFailed);
    }
    
    function create(destination) {
      CreateDestinationService(destination.name, destination.endpoint, 
         destination.username, destination.password)
          .success(createDestinationSuccess)
          .error(createDestinationFailed);
    }
        
    function createDestinationSuccess(data, status) {
      console.log('Successful created destination.');
      vm.reload();
    }
    
    function createDestinationFailed(data, status) {
      if(status === 409) {
        alert($filter('tr')('destination_already_exists', []));
      }
      console.log('Failed create destination:' + data);
    }
    
    function update(destination) {
      UpdateDestinationService(vm.targetId, destination)
        .success(updateDestinationSuccess)
        .error(updateDestinationFailed);
    }
    
    function updateDestinationSuccess(data, status) {
      console.log('Successful update destination.');
      vm.reload();
    }
    
    function updateDestinationFailed(data, status) {
      console.log('Failed update destination.');
    }
    
    
    function getDestinationSuccess(data, status) {
      var destination = data;
      vm0.name = destination.name;
      vm0.endpoint = destination.endpoint;
      vm0.username = destination.username;
      vm0.password = destination.password;
      
      ListDestinationPolicyService(destination.id)
        .success(listDestinationPolicySuccess)
        .error(listDestinationPolicyFailed);
    }
    
    function getDestinationFailed(data, status) {
      console.log('Failed get destination.');
    }
    
    function listDestinationPolicySuccess(data, status) {
      for(var i in data) {
        if(data[i].enabled === 1) {
          vm.editable = false;
          break;
        }
      }
    }
    
    function listDestinationPolicyFailed(data, status) {
      console.log('Failed list destination policy:' + data);
    }
    
    function pingDestination() {
      vm.pingAvailable = false;
      
      var target = {
        'name': vm0.name,
        'endpoint': vm0.endpoint,
        'username': vm0.username,
        'password': vm0.password
      };
      PingDestinationService(target)
        .success(pingDestinationSuccess)
        .error(pingDestinationFailed);
    }
    function pingDestinationSuccess(data, status) {
      vm.pingAvailable = true;
      vm.pingMessage = $filter('tr')('successful_ping_target', []);
    }
    function pingDestinationFailed(data, status) {
      vm.pingAvailable = true;
      vm.pingMessage = $filter('tr')('failed_ping_target', []) + (data && data.length > 0 ? ':' + data : '.');
    }
  }
  
  function createDestination() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/system-management/create-destination.directive.html',
      'scope': {
        'action': '@',
        'targetId': '@',
        'reload': '&'
      },
      'link': link,
      'controller': CreateDestinationController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      
      element.find('#createDestinationModal').on('show.bs.modal', function() {
        
        scope.form.$setPristine();
        scope.form.$setUntouched();
        
        ctrl.notAvailble = true;
        ctrl.pingAvailable = true;
        ctrl.pingMessage = '';
        
        switch(ctrl.action) {
        case 'ADD_NEW':
          ctrl.addNew();
          break;
        case 'EDIT':
          ctrl.edit(ctrl.targetId);
          break;
        }
        scope.$apply();
      });
      
      ctrl.save = save;
      
      function save(destination) {
        if(destination) {          
          switch(ctrl.action) {
          case 'ADD_NEW':
            ctrl.create(destination);
            break;
          case 'EDIT':
            ctrl.update(destination);
            break;
          }
          element.find('#createDestinationModal').modal('hide');
        }
      }
    }
  }
  
})();