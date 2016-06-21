(function() {
  
  'use strict';
  
  angular
    .module('harbor.system.management')
    .directive('createDestination', createDestination);
    
  CreateDestinationController.$inject = ['$scope', 'ListDestinationService', 'CreateDestinationService', 'UpdateDestinationService', 'PingDestinationService'];
  
  function CreateDestinationController($scope, ListDestinationService, CreateDestinationService, UpdateDestinationService, PingDestinationService) {
    var vm = this;
    
    $scope.destination = {};
    
    var vm0 = $scope.destination;
    vm.addNew = addNew;
    vm.edit = edit;
    vm.create = create;
    vm.update = update;
    vm.pingDestination = pingDestination;
    
    $scope.$watch('vm.action+","+vm.targetId', function(current) {
      if(current) {
        var parts = current.split(',');
        vm.action = parts[0];
        vm.targetId = parts[1];
        switch(vm.action) {
        case 'ADD_NEW':
          vm.modalTitle = 'Create destination';
          vm.addNew();
          break;
        case 'EDIT':
          vm.modalTitle = 'Edit destination';
          vm.edit(vm.targetId);
          break;
        }
      }      
    });
    
    function addNew() {
      vm0.name = '';
      vm0.endpoint = '';
      vm0.username = '';
      vm0.password = '';
    }
    
    function edit(targetId) {
      getDestination(targetId);
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
        alert('Destination already exists.');
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
    
    function getDestination(targetId) {
      ListDestinationService(targetId)
        .success(getDestinationSuccess)
        .error(getDestinationFailed);
    }
    
    function getDestinationSuccess(data, status) {
      var destination = data;
      vm0.name = destination.name;
      vm0.endpoint = destination.endpoint;
      vm0.username = destination.username;
      vm0.password = destination.password;
    }
    
    function getDestinationFailed(data, status) {
      console.log('Failed get destination.');
    }
    
    function pingDestination() {
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
      alert('Successful ping target.');
    }
    function pingDestinationFailed(data, status) {
      alert('Failed ping target:' + data);
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