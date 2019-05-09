*** Settings ***
Documentation  This resource provides any keywords related to the Harbor private registry appliance

*** Variables ***
${project_chart_tabpage}  xpath=//project-detail//a[contains(.,'Charts')]
${project_chart_list}  xpath=//hbr-helm-chart
${upload_chart_button}  //button[contains(.,'Upload')]
${chart_file_browse}  //*[@id='chart']
${chart_prov_browse}  //*[@id='prov']
${upload_action_button}  //clr-modal//form/div/button[contains(.,'Upload')]

${harbor_chart_name}  harbor
${harbor_chart_filename}  harbor-0.2.0.tgz
${harbor_chart_version}  0.2.0
${harbor_chart_prov_filename}  harbor-0.2.0.tgz.prov
${harbor_chart_file_url}  https://storage.googleapis.com/harbor-builds/helm-chart-test-files/harbor-0.2.0.tgz
${harbor_chart_prov_file_url}  https://storage.googleapis.com/harbor-builds/helm-chart-test-files/harbor-0.2.0.tgz.prov

${prometheus_chart_name}  prometheus
${prometheus_chart_filename}  prometheus-7.0.2.tgz
${prometheus_chart_version}  7.0.2
${prometheus_chart_file_url}  https://storage.googleapis.com/harbor-builds/helm-chart-test-files/prometheus-7.0.2.tgz
${prometheus_version}  //hbr-helm-chart//a[contains(.,'prometheus')]

${chart_detail}  //hbr-chart-detail
${summary_markdown}  //*[@id='summary-content']//div[contains(@class,'md-div')]
${summary_container}  //*[@id='summary-content']//div[contains(@class,'summary-container')]
${detail_dependency}  //*[@id='depend-link']
${dependency_content}  //*[@id='depend-content']/hbr-chart-detail-dependency
${detail_value}  //*[@id='value-link']
${value_content}  //*[@id='value-content']/hbr-chart-detail-value

${version_bread_crumbs}  //project-chart-detail//a[contains(.,'Versions')]
${version_checkbox}  //clr-dg-row//clr-checkbox-wrapper/label
${version_delete}  //clr-dg-action-bar/button[contains(.,'DELETE')]
${version_confirm_delete}  //clr-modal//button[contains(.,'DELETE')]

${helmchart_content}  //project-detail/project-list-charts/hbr-helm-chart