(function() {
  
  'use strict';
  
  angular
    .module('harbor.summary')
    .constant('projectStatistics', projectStatistics)
    .factory('getStatisticsName', getStatisticsName);
    
  function projectStatistics() {
    return [
      {'name': 'projects', 'payloadName': 'my_project_count'},
      {'name': 'repositories', 'payloadName': 'my_repo_count'},
      {'name': 'public_projects', 'payloadName': 'public_project_count'},
      {'name': 'public_repositories', 'payloadName': 'public_repo_count'},
      {'name': 'total_projects', 'payloadName': 'total_project_count'},
      {'name': 'total_repositories', 'payloadName': 'total_repo_count'},
    ];
  }
  
  getStatisticsName.$inject = ['projectStatistics'];
  
  function getStatisticsName(projectStatistics) {
    var r = projectStatistics();
    return get;     
    function get(query) {
     
      for(var i = 0; i < r.length; i++) {
        var StatisticsName = r[i];
        if(query.key === 'payloadName' && StatisticsName.payloadName === query.value
          || query.key === 'name' && StatisticsName.name === query.value) {
           return StatisticsName;
        }
      }
    }
  }
})();