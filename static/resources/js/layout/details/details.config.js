(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .filter('name', nameFilter);
    
  function nameFilter() {
   
    return filter;

    function filter(input, filterInput, key) {
      input = input || [];
      var filteredResults = [];
 
      if (filterInput !== '') {
        for(var i = 0; i < input.length; i++) {
          var item = input[i];
          if((key === "" && item.indexOf(filterInput) >= 0) || (key !== "" && item[key].indexOf(filterInput) >= 0)) {
            filteredResults.push(item);
            continue;
          }   
        }
        input = filteredResults;
      }
      return input;
    }
  }

  
})();