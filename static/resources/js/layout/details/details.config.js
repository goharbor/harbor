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