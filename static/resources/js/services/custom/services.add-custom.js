(function() {
  'use strict';

  angular
    .module('harbor.services.custom')
    .factory('AddCustomService', AddCustomService);

  AddCustomService.$inject = ['$http', '$log'];

  function AddCustomService($http, $log) {

    return AddCustom;

    function AddCustom(customName, tagName) {
      //创建客户
      return $http
        .post('/api/customer', {
          name : customName,
          tag : tagName
        });
    }
  }

})();
