(function() {

  'use strict';

  angular
    .module('harbor.services.custom')
    .factory('DeleteCustomService', DeleteCustomService);

  DeleteCustomService.$inject = ['$http', '$log'];

  function DeleteCustomService($http, $log) {

    return DeleteCustom;

    function DeleteCustom(customId) {
      return $http
        .delete('/api/customer/'+customId, {});
    }
  }

})();
