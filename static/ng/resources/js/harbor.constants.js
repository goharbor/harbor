(function() {
  'use strict';
  
  angular
    .module('harbor.app') 
    .constant('navigationTabs', navigationTabs);
  
  function navigationTabs() {
    var data = [
      {name: "Dashboard", url: "/ng/dashboard"},
      {name: "My Projects", url: "/ng/project"}];
    return data;
  }    
  
})();