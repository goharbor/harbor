/*
    Copyright (c) 2016 VMware, Inc. All Rights Reserved.
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
        
        http://www.apache.org/licenses/LICENSE-2.0
        
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/
(function() {
  
  'use strict';
  
  angular
    .module('harbor.system.management')
    .constant('defaultPassword', '12345678')
    .directive('configuration', configuration);
  
  ConfigurationController.$inject = ['$scope', 'ConfigurationService', 'defaultPassword', '$filter', 'trFilter'];
  
  function ConfigurationController($scope, ConfigurationService, defaultPassword, $filter, trFilter) {
    var vm = this;
    
    vm.toggleBooleans = [
      {
        'name': 'True',
        'value': true
      },
      {
        'name': 'False',
        'value': false
      }
    ];
    
    vm.toggleCustoms = [
      {
        'name': 'Admin Only',
        'value': 'adminonly',
      },
      {
        'name': 'Everyone',
        'value': 'everyone'
      }
    ];
    
    vm.supportedAuths = [
      {
        'name': 'DB auth',
        'value': 'db_auth'
      },
      {
        'name': 'LDAP auth',
        'value': 'ldap_auth'
      }
    ];

    var confKeyDefinitions = {
      'auth_mode': { type: 'auth', attr: 'authMode' },
      'self_registration': { type: 'auth', attr: 'selfRegistration' },
      'ldap_url': { type: 'auth', attr: 'ldapURL' },
      'ldap_search_dn': { type: 'auth', attr: 'ldapSearchDN' },
      'ldap_search_password': { type: 'auth', attr: 'ldapSearchPassword' },
      'ldap_base_dn': { type: 'auth', attr: 'ldapBaseDN' },
      'ldap_uid': { type: 'auth', attr: 'ldapUID' },
      'ldap_filter': { type: 'auth', attr: 'ldapFilter' },
      'ldap_timeout': { type: 'auth', attr: 'ldapConnectionTimeout' },
      'ldap_scope': { type: 'auth', attr: 'ldapScope' },
      'email_host': { type: 'email', attr: 'server' },
      'email_port': { type: 'email', attr: 'serverPort' },
      'email_username': { type: 'email', attr: 'username' },
      'email_password': { type: 'email', attr: 'password' },
      'email_from': { type: 'email', attr: 'from' },
      'email_ssl': { type: 'email', attr: 'SSL' },
      'project_creation_restriction': { type: 'system', attr: 'projectCreationRestriction' },
      'verify_remote_cert': { type: 'system', attr: 'verifyRemoteCert' }
    };
    
    $scope.auth   = {};
    $scope.email  = {};
    $scope.system = {};
    
    vm.retrieve = retrieve;
    
    vm.saveAuthConf = saveAuthConf;
    vm.saveEmailConf = saveEmailConf;
    vm.saveSystemConf = saveSystemConf;
    
    vm.gatherUpdateItems = gatherUpdateItems;
    vm.clearUp = clearUp;
    vm.hasChanged = hasChanged;
    vm.setMaskPassword = setMaskPassword;   
    vm.undo = undo;
    
    vm.pingLDAP = pingLDAP;
    vm.pingTIP = false;
    vm.isError = false;
    vm.pingMessage = '';
    
    vm.retrieve();
    
    function retrieve() {                

      vm.ldapSearchPasswordChanged = false;
      vm.emailPasswordChanged = false;  
      vm.changedItems = {};
      vm.updatedItems = {};
      vm.warning = {};
      vm.editable = {};
      
      ConfigurationService
        .get()
        .then(getConfigurationSuccess, getConfigurationFailed);
    }

    function getConfigurationSuccess(response) {
      var data = response.data || [];
      for(var key in data) {
        var mappedDef = keyMapping(key);
        if(mappedDef) {
          $scope[mappedDef['type']][mappedDef['attr']] = { 'target': mappedDef['type'] + '.' + mappedDef['attr'], 'data': valueMapping(data[key]['value']) };
          $scope.$watch(mappedDef['type'] + '.' + mappedDef['attr'], onChangedCallback, true);
          $scope[mappedDef['type']][mappedDef['attr']]['origin'] = { 'target': mappedDef['type'] + '.' + mappedDef['attr'], 'data': valueMapping(data[key]['value']) };
          vm.editable[mappedDef['type'] + '.' + mappedDef['attr']] = data[key]['editable'];
        }
      }
      
      $scope.auth.ldapSearchPassword = { 'target': 'auth.ldapSearchPassword', 'data': defaultPassword};
      $scope.email.password = { 'target': 'email.password', 'data': defaultPassword};
            
      $scope.$watch('auth.ldapSearchPassword', onChangedCallback, true);
      $scope.$watch('email.password', onChangedCallback, true);
      
      $scope.auth.ldapSearchPassword.actual = { 'target': 'auth.ldapSearchPassword', 'data': ''};
      $scope.email.password.actual = { 'target': 'email.password', 'data': ''};
    }
                   
    function keyMapping(confKey) {
      for (var key in confKeyDefinitions) {
        if (confKey === key) {
          return confKeyDefinitions[key];
        } 
      }
      return null;
    }
    
    function valueMapping(value) {
      switch(value) {
      case true:
        return vm.toggleBooleans[0];
      case false:
        return vm.toggleBooleans[1];
      case 'db_auth':
        return vm.supportedAuths[0];
      case 'ldap_auth':
        return vm.supportedAuths[1];
      case 'adminonly':
        return vm.toggleCustoms[0];
      case 'everyone':
        return vm.toggleCustoms[1];
      default: 
        return value;
      }
    }
        
    function onChangedCallback(current, previous) {
      if(!angular.equals(current, previous)) {
        var compositeKey = current.target.split(".");
        vm.changed = false;
        var changedData = {};
        switch(current.target) {
        case 'auth.ldapSearchPassword':
          if(vm.ldapSearchPasswordChanged) {
            vm.changed = true;
            changedData = $scope.auth.ldapSearchPassword.actual.data;
          }
          break;
        case 'email.password':
          if(vm.emailPasswordChanged) {
            vm.changed = true;
            changedData = $scope.email.password.actual.data;
          }
          break;
        default:
          if(!angular.equals(current.data, $scope[compositeKey[0]][compositeKey[1]]['origin']['data'])) {
            vm.changed = true;    
            changedData = current.data;
          }  
        }
        if(vm.changed) {
          vm.changedItems[current.target] = changedData;
          vm.warning[current.target] = true;
        } else {
          delete vm.changedItems[current.target];
          vm.warning[current.target] = false;
        }
      }
    }   
    
    function getConfigurationFailed(response) {
      console.log('Failed to get configurations.');
    }

    function updateConfigurationSuccess(response) {
      $scope.$emit('modalTitle', $filter('tr')('update_configuration_title', []));
      $scope.$emit('modalMessage',  $filter('tr')('successful_update_configuration', []));      
      var emitInfo = {
        'confirmOnly': true,
        'contentType': 'text/plain',
        'action' : function() {
          vm.retrieve();
        }
      };      
      $scope.$emit('raiseInfo', emitInfo);      
      console.log('Updated system configuration successfully.');
    }
    
    function updateConfigurationFailed() {
      $scope.$emit('modalTitle', $filter('tr')('update_configuration_title', []));
      $scope.$emit('modalMessage',  $filter('tr')('failed_to_update_configuration', []));      
      $scope.$emit('raiseError', true);
      console.log('Failed to update system configurations.');
    }

    function gatherUpdateItems() {     
      vm.updatedItems = {};
      for(var key in confKeyDefinitions) {
        var value = confKeyDefinitions[key];
        var compositeKey = value.type + '.' + value.attr;
        for(var itemKey in vm.changedItems) {
          var item = vm.changedItems[itemKey];
          if (compositeKey === itemKey) {
            (typeof item === 'object' && item) ? vm.updatedItems[key] = ((typeof item.value === 'boolean') ? Number(item.value) + '' : item.value) : vm.updatedItems[key] = String(item) || '';
          }
        }
      }
    }
        
    function saveAuthConf(auth) {
      vm.gatherUpdateItems();
      console.log('auth changed:' + angular.toJson(vm.updatedItems));
      ConfigurationService
        .update(vm.updatedItems)
        .then(updateConfigurationSuccess, updateConfigurationFailed);
    }
    
    function saveEmailConf(email) {
      vm.gatherUpdateItems();
      console.log('email changed:' + angular.toJson(vm.updatedItems));
      ConfigurationService
        .update(vm.updatedItems)
        .then(updateConfigurationSuccess, updateConfigurationFailed);
    }
    
    function saveSystemConf(system) {
      vm.gatherUpdateItems();
      console.log('system changed:' + angular.toJson(vm.updatedItems));
      ConfigurationService
        .update(vm.updatedItems)
        .then(updateConfigurationSuccess, updateConfigurationFailed);
    }
    
    function clearUp(input) {
      switch(input.target) {
      case 'auth.ldapSearchPassword':
        $scope.auth.ldapSearchPassword.data = '';
        break;
      case 'email.password':
        $scope.email.password.data = '';
        break;
      }
    }
    
    function hasChanged(input) {
      switch(input.target) {
      case 'auth.ldapSearchPassword':
        vm.ldapSearchPasswordChanged = true;
        $scope.auth.ldapSearchPassword.actual.data = input.data;
        break;
      case 'email.password':
        vm.emailPasswordChanged = true;  
        $scope.email.password.actual.data = input.data;
        break;
      }
    }
    
    function setMaskPassword(input) {
      input.data = defaultPassword;
    }
    
    function undo() {
      vm.retrieve();
    }
   
    function pingLDAP(auth) {
      var keyset = [
        {'name': 'ldapURL'     , 'attr': 'ldap_url'}, 
        {'name': 'ldapSearchDN', 'attr': 'ldap_search_dn'},
        {'name': 'ldapScope'   , 'attr': 'ldap_scope'},
        {'name': 'ldapSearchPassword'   , 'attr': 'ldap_search_password'},
        {'name': 'ldapConnectionTimeout', 'attr': 'ldap_connection_timeout'}
      ];
      var ldapConf = {};
      
      for(var i = 0; i < keyset.length; i++) {
        var key = keyset[i];
        var value;
        if(key.name === 'ldapSearchPassword') {
          value = auth[key.name]['actual']['data'];
        }else {
          value = auth[key.name]['data'];  
        }
        ldapConf[key.attr] = value;
      }
      
      vm.pingMessage = $filter('tr')('pinging_target');
      vm.pingTIP = true;
      vm.isError = false;
           
      ConfigurationService
        .pingLDAP(ldapConf)
        .then(pingLDAPSuccess, pingLDAPFailed);
    }
    
    function pingLDAPSuccess(response) {
      vm.pingTIP = false;
      vm.pingMessage = $filter('tr')('successful_ping_target');
    }
    
    function pingLDAPFailed(response) {
      vm.isError = true;
      vm.pingTIP = false;
      vm.pingMessage = $filter('tr')('failed_to_ping_target');
      console.log('Failed to ping LDAP target:' + response.data);
    }
   
  }
  
  configuration.$inject = ['$filter', 'trFilter'];
  
  function configuration($filter, trFilter) {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/system-management/configuration.directive.html',
      'scope': true,
      'link': link,
      'controller': ConfigurationController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      element.find('#ulTabHeader a').on('click', function(e) {
        e.preventDefault();
        ctrl.gatherUpdateItems();
        if(!angular.equals(ctrl.updatedItems, {})) {
          var emitInfo = {
           'confirmOnly': true,
           'contentType': 'text/html',
           'action' : function() {
              return;  
            }
          };
          scope.$emit('modalTitle', $filter('tr')('caution'));
          scope.$emit('modalMessage', $filter('tr')('please_save_changes'));
          scope.$emit('raiseInfo', emitInfo);
          scope.$apply();
          e.stopPropagation();
        }else{
          $(this).tab('show');
        }
        
      });
      element.find('#ulTabHeader a:first').trigger('click');
    }
  }
  
})();