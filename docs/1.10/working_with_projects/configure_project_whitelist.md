[Back to table of contents](../index.md)

----------

# Configure a Per-Project CVE Whitelist

When you run vulnerability scans, images that are subject to Common Vulnerabilities and Exposures (CVE) are identified. According to the severity of the CVE and your security settings, these images might not be permitted to run. You can create whitelists of CVEs to ignore during vulnerability scanning. 

Harbor administrators can set a system-wide CVE whitelist. For information about site-wide CVE whitelists, see [Configure System-Wide CVE Whitelists](../administration/vulnerability_scanning/configire_system_whitelist.md). By default, the system whitelist is applied to all projects. You can configure different CVE whitelists for individual projects, that override the system whitelist. 

1. Go to **Projects**, select a project, and select **Configuration**.
1. Under **CVE whitelist**, select **Project whitelist**. 
   ![Project CVE whitelist](../img/cve-whitelist5.png)
1. Optionally click **Copy From System** to add all of the CVE IDs from the system CVE whitelist to this project whitelist.
1. Click **Add** and enter a list of additional CVE IDs to ignore during vulnerability scanning of this project. 
   ![Add project CVEs](../img/cve-whitelist6.png)

   Either use a comma-separated list or newlines to add multiple CVE IDs to the list.
1. Click **Add** at the bottom of the window to add the CVEs to the project whitelist.
1. Optionally uncheck the **Never expires** checkbox and use the calendar selector to set an expiry date for the whitelist.
1. Click **Save** at the bottom of the page to save your settings.

After you have created a project whitelist, you can remove CVE IDs from the list by clicking the delete button next to it in the list. You can click **Add** at any time to add more CVE IDs to this project whitelist. 

If CVEs are added to the system whitelist after you have created a project whitelist, click **Copy From System** to add the new entries from the system whitelist to the project whitelist. 

**NOTE**: If CVEs are deleted from the system whitelist after you have created a project whitelist, and if you added the system whitelist to the project whitelist, you must manually remove the deleted CVEs from the project whitelist. If you click **Copy From System** after CVEs have been deleted from the system whitelist, the deleted CVEs are not automatically removed from the project whitelist.

----------

[Back to table of contents](../index.md)
