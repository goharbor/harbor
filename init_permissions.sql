DO $$
DECLARE
	ResourceSelf 			  VARCHAR(50) := '';
	ResourceMember 			   VARCHAR(50) := 'member';
	ResourceMetadata 		   VARCHAR(50) := 'metadata';
	ResourceLog   			   VARCHAR(50) := 'log';
	ResourceLabel              VARCHAR(50) := 'label';
	ResourceQuota              VARCHAR(50) := 'quota';
	ResourceRepository         VARCHAR(50) := 'repository';
	ResourceTagRetention       VARCHAR(50) := 'tag-retention';
	ResourceImmutableTag       VARCHAR(50) := 'immutable-tag';
	ResourceConfiguration      VARCHAR(50) := 'configuration';
	ResourceRobot              VARCHAR(50) := 'robot';
	ResourceNotificationPolicy VARCHAR(50) := 'notification-policy';
	ResourceScan               VARCHAR(50) := 'scan';
	ResourceSBOM               VARCHAR(50) := 'sbom';
	ResourceScanner            VARCHAR(50) := 'scanner';
	ResourceArtifact           VARCHAR(50) := 'artifact';
	ResourceArtifactAddition   VARCHAR(50) := 'artifact-addition';
	ResourceTag                VARCHAR(50) := 'tag';
	ResourceAccessory          VARCHAR(50) := 'accessory';
	ResourceArtifactLabel      VARCHAR(50) := 'artifact-label';
	ResourcePreatPolicy        VARCHAR(50) := 'preheat-policy';
	ResourceExportCVE          VARCHAR(50) := 'export-cve';
	
	ActionCreate 				VARCHAR(50) := 'create';
	ActionRead 					VARCHAR(50) := 'read';
	ActionUpdate 				VARCHAR(50) := 'update';
	ActionDelete 				VARCHAR(50) := 'delete';
	ActionList 					VARCHAR(50) := 'list';
	ActionPull 					VARCHAR(50) := 'pull';
	ActionPush 					VARCHAR(50) := 'push';
	ActionOperate     			VARCHAR(50) := 'operate';
	ActionStop        			VARCHAR(50) := 'stop';
	
	PROJECT_ADMIN VARCHAR(50) := 'projectAdmin';
	MAINTAINER VARCHAR(50) := 'maintainer';
	DEVELOPER VARCHAR(50) := 'developer';
	GUEST VARCHAR(50) :=	'guest';
	LIMITED_GUEST VARCHAR(50) := 'limitedGuest';
	
	
BEGIN

	-- Temporary table 
	
	CREATE TEMPORARY TABLE "#perms" (resource VARCHAR(50), action VARCHAR(50), role_name VARCHAR(50)); 
	INSERT INTO "#perms" (resource, action, role_name) VALUES 
	(ResourceSelf, ActionRead, PROJECT_ADMIN), 
	(ResourceSelf, ActionUpdate, PROJECT_ADMIN), 
	(ResourceSelf, ActionDelete, PROJECT_ADMIN), 

	(ResourceMember, ActionCreate, PROJECT_ADMIN), 
	(ResourceMember, ActionRead, PROJECT_ADMIN), 
	(ResourceMember, ActionUpdate, PROJECT_ADMIN), 
	(ResourceMember, ActionDelete, PROJECT_ADMIN), 
	(ResourceMember, ActionList, PROJECT_ADMIN), 

	(ResourceMetadata, ActionCreate, PROJECT_ADMIN), 
	(ResourceMetadata, ActionRead, PROJECT_ADMIN), 
	(ResourceMetadata, ActionUpdate, PROJECT_ADMIN), 
	(ResourceMetadata, ActionDelete, PROJECT_ADMIN), 

	(ResourceLog, ActionList, PROJECT_ADMIN), 

	(ResourceLabel, ActionCreate, PROJECT_ADMIN), 
	(ResourceLabel, ActionRead, PROJECT_ADMIN), 
	(ResourceLabel, ActionUpdate, PROJECT_ADMIN), 
	(ResourceLabel, ActionDelete, PROJECT_ADMIN), 
	(ResourceLabel, ActionList, PROJECT_ADMIN), 

	(ResourceQuota, ActionRead, PROJECT_ADMIN), 

	(ResourceRepository, ActionCreate, PROJECT_ADMIN), 
	(ResourceRepository, ActionRead, PROJECT_ADMIN), 
	(ResourceRepository, ActionUpdate, PROJECT_ADMIN), 
	(ResourceRepository, ActionDelete, PROJECT_ADMIN), 
	(ResourceRepository, ActionList, PROJECT_ADMIN), 
	(ResourceRepository, ActionPull, PROJECT_ADMIN), 
	(ResourceRepository, ActionPush, PROJECT_ADMIN), 

	(ResourceTagRetention, ActionCreate, PROJECT_ADMIN), 
	(ResourceTagRetention, ActionRead, PROJECT_ADMIN), 
	(ResourceTagRetention, ActionUpdate, PROJECT_ADMIN), 
	(ResourceTagRetention, ActionDelete, PROJECT_ADMIN), 
	(ResourceTagRetention, ActionList, PROJECT_ADMIN), 
	(ResourceTagRetention, ActionOperate, PROJECT_ADMIN), 

	(ResourceImmutableTag, ActionCreate, PROJECT_ADMIN), 
	(ResourceImmutableTag, ActionUpdate, PROJECT_ADMIN), 
	(ResourceImmutableTag, ActionDelete, PROJECT_ADMIN), 
	(ResourceImmutableTag, ActionList, PROJECT_ADMIN), 

	(ResourceConfiguration, ActionRead, PROJECT_ADMIN), 
	(ResourceConfiguration, ActionUpdate, PROJECT_ADMIN), 

	(ResourceRobot, ActionCreate, PROJECT_ADMIN), 
	(ResourceRobot, ActionRead, PROJECT_ADMIN), 
	(ResourceRobot, ActionUpdate, PROJECT_ADMIN), 
	(ResourceRobot, ActionDelete, PROJECT_ADMIN), 
	(ResourceRobot, ActionList, PROJECT_ADMIN), 

	(ResourceNotificationPolicy, ActionCreate, PROJECT_ADMIN), 
	(ResourceNotificationPolicy, ActionUpdate, PROJECT_ADMIN), 
	(ResourceNotificationPolicy, ActionDelete, PROJECT_ADMIN), 
	(ResourceNotificationPolicy, ActionList, PROJECT_ADMIN), 
	(ResourceNotificationPolicy, ActionRead, PROJECT_ADMIN), 

	(ResourceScan, ActionCreate, PROJECT_ADMIN), 
	(ResourceScan, ActionRead, PROJECT_ADMIN), 
	(ResourceScan, ActionStop, PROJECT_ADMIN), 
	(ResourceSBOM, ActionCreate, PROJECT_ADMIN), 
	(ResourceSBOM, ActionStop, PROJECT_ADMIN), 
	(ResourceSBOM, ActionRead, PROJECT_ADMIN), 

	(ResourceScanner, ActionRead, PROJECT_ADMIN), 
	(ResourceScanner, ActionCreate, PROJECT_ADMIN), 

	(ResourceArtifact, ActionCreate, PROJECT_ADMIN), 
	(ResourceArtifact, ActionRead, PROJECT_ADMIN), 
	(ResourceArtifact, ActionDelete, PROJECT_ADMIN), 
	(ResourceArtifact, ActionList, PROJECT_ADMIN), 
	(ResourceArtifactAddition, ActionRead, PROJECT_ADMIN), 

	(ResourceTag, ActionList, PROJECT_ADMIN), 
	(ResourceTag, ActionCreate, PROJECT_ADMIN), 
	(ResourceTag, ActionDelete, PROJECT_ADMIN), 

	(ResourceAccessory, ActionList, PROJECT_ADMIN), 

	(ResourceArtifactLabel, ActionCreate, PROJECT_ADMIN), 
	(ResourceArtifactLabel, ActionDelete, PROJECT_ADMIN), 

	(ResourcePreatPolicy, ActionCreate, PROJECT_ADMIN), 
	(ResourcePreatPolicy, ActionRead, PROJECT_ADMIN), 
	(ResourcePreatPolicy, ActionUpdate, PROJECT_ADMIN), 
	(ResourcePreatPolicy, ActionDelete, PROJECT_ADMIN), 
	(ResourcePreatPolicy, ActionList, PROJECT_ADMIN), 

	(ResourceExportCVE, ActionCreate, PROJECT_ADMIN), 
	(ResourceExportCVE, ActionRead, PROJECT_ADMIN), 
	(ResourceExportCVE, ActionList, PROJECT_ADMIN), 

	(ResourceSelf, ActionRead, MAINTAINER), 

	(ResourceMember, ActionRead, MAINTAINER), 
	(ResourceMember, ActionList, MAINTAINER), 

	(ResourceMetadata, ActionRead, MAINTAINER), 

	(ResourceQuota, ActionRead, MAINTAINER), 

	(ResourceLabel, ActionCreate, MAINTAINER), 
	(ResourceLabel, ActionRead, MAINTAINER), 
	(ResourceLabel, ActionUpdate, MAINTAINER), 
	(ResourceLabel, ActionDelete, MAINTAINER), 
	(ResourceLabel, ActionList, MAINTAINER), 

	(ResourceRepository, ActionCreate, MAINTAINER), 
	(ResourceRepository, ActionRead, MAINTAINER), 
	(ResourceRepository, ActionUpdate, MAINTAINER), 
	(ResourceRepository, ActionDelete, MAINTAINER), 
	(ResourceRepository, ActionList, MAINTAINER), 
	(ResourceRepository, ActionPush, MAINTAINER), 
	(ResourceRepository, ActionPull, MAINTAINER), 

	(ResourceTagRetention, ActionCreate, MAINTAINER), 
	(ResourceTagRetention, ActionRead, MAINTAINER), 
	(ResourceTagRetention, ActionUpdate, MAINTAINER), 
	(ResourceTagRetention, ActionDelete, MAINTAINER), 
	(ResourceTagRetention, ActionList, MAINTAINER), 
	(ResourceTagRetention, ActionOperate, MAINTAINER), 

	(ResourceAccessory, ActionList, MAINTAINER), 

	(ResourceImmutableTag, ActionCreate, MAINTAINER), 
	(ResourceImmutableTag, ActionUpdate, MAINTAINER), 
	(ResourceImmutableTag, ActionDelete, MAINTAINER), 
	(ResourceImmutableTag, ActionList, MAINTAINER), 

	(ResourceConfiguration, ActionRead, MAINTAINER), 

	(ResourceRobot, ActionRead, MAINTAINER), 
	(ResourceRobot, ActionList, MAINTAINER), 

	(ResourceNotificationPolicy, ActionRead, MAINTAINER), 
	(ResourceNotificationPolicy, ActionList, MAINTAINER), 

	(ResourceScan, ActionCreate, MAINTAINER), 
	(ResourceScan, ActionRead, MAINTAINER), 
	(ResourceScan, ActionStop, MAINTAINER), 
	(ResourceSBOM, ActionCreate, MAINTAINER), 
	(ResourceSBOM, ActionStop, MAINTAINER), 
	(ResourceSBOM, ActionRead, MAINTAINER), 

	(ResourceScanner, ActionRead, MAINTAINER), 

	(ResourceArtifact, ActionCreate, MAINTAINER), 
	(ResourceArtifact, ActionRead, MAINTAINER), 
	(ResourceArtifact, ActionDelete, MAINTAINER), 
	(ResourceArtifact, ActionList, MAINTAINER), 
	(ResourceArtifactAddition, ActionRead, MAINTAINER), 

	(ResourceTag, ActionList, MAINTAINER), 
	(ResourceTag, ActionCreate, MAINTAINER), 
	(ResourceTag, ActionDelete, MAINTAINER), 

	(ResourceArtifactLabel, ActionCreate, MAINTAINER), 
	(ResourceArtifactLabel, ActionDelete, MAINTAINER), 

	(ResourceExportCVE, ActionCreate, MAINTAINER), 
	(ResourceExportCVE, ActionRead, MAINTAINER), 
	(ResourceExportCVE, ActionList, MAINTAINER), 
		
	(ResourceSelf, ActionRead, DEVELOPER), 

	(ResourceMember, ActionRead, DEVELOPER), 
	(ResourceMember, ActionList, DEVELOPER), 

	(ResourceLabel, ActionRead, DEVELOPER), 
	(ResourceLabel, ActionList, DEVELOPER), 

	(ResourceQuota, ActionRead, DEVELOPER), 

	(ResourceRepository, ActionCreate, DEVELOPER), 
	(ResourceRepository, ActionRead, DEVELOPER), 
	(ResourceRepository, ActionUpdate, DEVELOPER), 
	(ResourceRepository, ActionList, DEVELOPER), 
	(ResourceRepository, ActionPush, DEVELOPER), 
	(ResourceRepository, ActionPull, DEVELOPER), 

	(ResourceTagRetention, ActionCreate, DEVELOPER), 
	(ResourceTagRetention, ActionRead, DEVELOPER), 
	(ResourceTagRetention, ActionUpdate, DEVELOPER), 
	(ResourceTagRetention, ActionDelete, DEVELOPER), 
	(ResourceTagRetention, ActionList, DEVELOPER), 
	(ResourceTagRetention, ActionOperate, DEVELOPER), 

	(ResourceConfiguration, ActionRead, DEVELOPER), 

	(ResourceRobot, ActionRead, DEVELOPER), 
	(ResourceRobot, ActionList, DEVELOPER), 

	(ResourceScan, ActionRead, DEVELOPER), 
	(ResourceSBOM, ActionRead, DEVELOPER), 

	(ResourceScanner, ActionRead, DEVELOPER), 

	(ResourceArtifact, ActionCreate, DEVELOPER), 
	(ResourceArtifact, ActionRead, DEVELOPER), 
	(ResourceArtifact, ActionList, DEVELOPER), 
	(ResourceArtifactAddition, ActionRead, DEVELOPER), 

	(ResourceTag, ActionList, DEVELOPER), 
	(ResourceTag, ActionCreate, DEVELOPER), 

	(ResourceAccessory, ActionList, DEVELOPER), 

	(ResourceArtifactLabel, ActionCreate, DEVELOPER), 
	(ResourceArtifactLabel, ActionDelete, DEVELOPER), 

	(ResourceExportCVE, ActionCreate, DEVELOPER), 
	(ResourceExportCVE, ActionRead, DEVELOPER), 
	(ResourceExportCVE, ActionList, DEVELOPER), 

	(ResourceSelf, ActionRead, GUEST), 

	(ResourceMember, ActionRead, GUEST), 
	(ResourceMember, ActionList, GUEST), 

	(ResourceLabel, ActionRead, GUEST), 
	(ResourceLabel, ActionList, GUEST), 

	(ResourceQuota, ActionRead, GUEST), 

	(ResourceRepository, ActionRead, GUEST), 
	(ResourceRepository, ActionList, GUEST), 
	(ResourceRepository, ActionPull, GUEST), 

	(ResourceConfiguration, ActionRead, GUEST), 

	(ResourceRobot, ActionRead, GUEST), 
	(ResourceRobot, ActionList, GUEST), 

	(ResourceScan, ActionRead, GUEST), 
	(ResourceSBOM, ActionRead, GUEST), 

	(ResourceScanner, ActionRead, GUEST), 

	(ResourceTag, ActionList, GUEST), 
	(ResourceAccessory, ActionList, GUEST), 

	(ResourceArtifact, ActionRead, GUEST), 
	(ResourceArtifact, ActionList, GUEST), 
	(ResourceArtifactAddition, ActionRead, GUEST), 

	(ResourceSelf, ActionRead, LIMITED_GUEST), 

	(ResourceQuota, ActionRead, LIMITED_GUEST), 

	(ResourceRepository, ActionList, LIMITED_GUEST), 
	(ResourceRepository, ActionRead, LIMITED_GUEST), 
	(ResourceRepository, ActionPull, LIMITED_GUEST), 

	(ResourceConfiguration, ActionRead, LIMITED_GUEST), 

	(ResourceScan, ActionRead, LIMITED_GUEST), 
	(ResourceSBOM, ActionRead, LIMITED_GUEST), 

	(ResourceScanner, ActionRead, LIMITED_GUEST), 

	(ResourceTag, ActionList, LIMITED_GUEST), 
	(ResourceAccessory, ActionList, LIMITED_GUEST), 

	(ResourceArtifact, ActionRead, LIMITED_GUEST), 
	(ResourceArtifact, ActionList, LIMITED_GUEST), 
	(ResourceArtifactAddition, ActionRead, LIMITED_GUEST);


	RAISE NOTICE 'created temporary table #perms';
	
	DECLARE 
		Cur_1 CURSOR FOR SELECT resource, action, role_name FROM "#perms";
		tmp_resource VARCHAR(50);
		tmp_action VARCHAR(50);
		tmp_role_name VARCHAR(50);
		tmp_role_id INT;
		tmp_policy_id INT;
		tmp_permission_id INT;

	BEGIN
		
		FOR r IN Cur_1 LOOP

			RAISE NOTICE 'permision: resource=% - a=% - role=%', r.resource, r.action, r.role_name;
			
			SELECT role_id into tmp_role_id FROM role WHERE name = r.role_name;

			IF EXISTS(
					SELECT * FROM permission_policy 
					WHERE scope = 'project-role' and resource = r.resource and action = r.action) THEN
			
				BEGIN
					SELECT id into tmp_policy_id FROM permission_policy 
					WHERE scope = 'project-role' and resource = r.resource and action = r.action;	
					
					RAISE NOTICE 'policy exists : %', tmp_policy_id;
				END;
			
			ELSE
				BEGIN
				    
					INSERT INTO permission_policy ( scope, resource, action,  creation_time  ) VALUES 
					('project-role', r.resource, r.action,  NOW()) RETURNING id INTO tmp_policy_id;
					
					RAISE NOTICE 'created policy : %', tmp_policy_id;
				END;

			END IF; 


			IF NOT EXISTS(
					SELECT * FROM role_permission 
					WHERE role_type = 'project-role' and role_id = tmp_role_id and permission_policy_id = tmp_policy_id) THEN
			
				BEGIN

					INSERT INTO role_permission (role_type, role_id, permission_policy_id, creation_time) VALUES
						('project-role', tmp_role_id, tmp_policy_id, NOW())  RETURNING id INTO tmp_permission_id;

					RAISE NOTICE 'created permission : %', tmp_permission_id;

				END;
			
			END IF;

		END LOOP;


	END;

END $$;