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
    .module('harbor.optional.menu')
    .directive('optionalMenu', optionalMenu);

  OptionalMenuController.$inject = ['$scope', '$window', 'I18nService', 'LogOutService', 'currentUser', '$timeout', 'trFilter', '$filter', 'GetVolumeInfoService'];

  function OptionalMenuController($scope, $window, I18nService, LogOutService, currentUser, $timeoutm, trFilter, $filter, GetVolumeInfoService) {
    var vm = this;
    
    var i18n = I18nService();
    i18n.setCurrentLanguage(vm.language);
    vm.languageName = i18n.getLanguageName(vm.language);
    console.log('current language:' + vm.languageName);
    
    vm.supportLanguages = i18n.getSupportLanguages(); 
    vm.user = currentUser.get();
    vm.setLanguage = setLanguage;     
    vm.logOut = logOut;
    vm.about = about;
            
    function setLanguage(language) {
      vm.languageName = i18n.getLanguageName(vm.language);
      var hash = $window.location.hash;
      $window.location.href = '/language?lang=' + language + '&hash=' + encodeURIComponent(hash);    
    }
    function logOut() {
      LogOutService()
        .success(logOutSuccess)
        .error(logOutFailed);
    }
    function logOutSuccess(data, status) {
      currentUser.unset();
      $window.location.href= '/';
    }
    function logOutFailed(data, status) {
      console.log('Failed to log out:' + data);
    }
    
    var raiseInfo = {
      'confirmOnly': true,
      'contentType': 'text/html',
      'action': function() {}
    };
      
    function about() {
      $scope.$emit('modalTitle', $filter('tr')('about_harbor'));
      vm.modalMessage = $filter('tr')('current_version', [vm.version || 'Unknown']);  
      if(vm.showDownloadCert === 'true') {
        appendDownloadCertLink();
      }
      GetVolumeInfoService("data")
        .then(getVolumeInfoSuccess, getVolumeInfoFailed);
    }
    function getVolumeInfoSuccess(response) {
      var storage = response.data;
      vm.modalMessage += '<br/>' + $filter('tr')('current_storage',
        [toGigaBytes(storage['storage']['free']), toGigaBytes(storage['storage']['total'])]);
      $scope.$emit('modalMessage', vm.modalMessage);
      $scope.$emit('raiseInfo', raiseInfo);
      
    }
    function getVolumeInfoFailed(response) {
      $scope.$emit('modalMessage', vm.modalMessage);
      $scope.$emit('raiseInfo', raiseInfo);
    }
    
    function toGigaBytes(val) {
      return Math.round(val / (1024 * 1024 * 1024));
    }
    
    function appendDownloadCertLink() {    
      vm.modalMessage += '<br/>' + $filter('tr')('default_root_cert', ['/api/systeminfo/getcert', $filter('tr')('download')]);
    }
    
  }
  
  function optionalMenu() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/optional_menu?timestamp=' + new Date().getTime(),
      'scope': {
        'version': '@',
        'language': '@',
        'showDownloadCert': '@'
      },
      'controller': OptionalMenuController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
  
})();