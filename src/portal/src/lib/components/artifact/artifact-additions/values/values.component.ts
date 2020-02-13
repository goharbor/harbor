import {
  Component,
  Input,
  OnInit,
} from "@angular/core";
import { ErrorHandler } from "../../../../utils/error-handler";
import { AdditionsService } from "../additions.service";
import { AdditionLink } from "../../../../../../ng-swagger-gen/models/addition-link";

@Component({
  selector: "hbr-artifact-values",
  templateUrl: "./values.component.html",
  styleUrls: ["./values.component.scss"],
})
export class ValuesComponent implements OnInit {
  @Input()
  valuesLink: AdditionLink;

  values: any = {
    "adminserver.image.pullPolicy": "IfNotPresent",
    "adminserver.image.repository": "vmware/harbor-adminserver",
    "adminserver.image.tag": "dev",
    "adminserver.tolerations": [],
    "adminserver.volumes.config.accessMode": "ReadWriteOnce",
    "adminserver.volumes.config.size": "1Gi",
    "authenticationMode": "db_auth",
    "chartmuseum.enabled": true,
    "chartmuseum.image.pullPolicy": "IfNotPresent",
    "chartmuseum.image.repository": "vmware/chartmuseum-photon",
    "chartmuseum.image.tag": "dev",
    "chartmuseum.tolerations": [],
    "chartmuseum.volumes.data.accessMode": "ReadWriteOnce",
    "chartmuseum.volumes.data.size": "5Gi",
    "clair.enabled": true,
    "clair.image.pullPolicy": "IfNotPresent",
    "clair.image.repository": "vmware/clair-photon",
    "clair.image.tag": "dev",
    "clair.tolerations": [],
    "clair.volumes.pgData.accessMode": "ReadWriteOnce",
    "clair.volumes.pgData.size": "1Gi",
    "database.external.clairDatabase": "clair",
    "database.external.coreDatabase": "registry",
    "database.external.host": "192.168.0.1",
    "database.external.notaryServerDatabase": "notary_server",
    "database.external.notarySignerDatabase": "notary_signer",
    "database.external.password": "password",
    "database.external.port": "5432",
    "database.external.username": "user",
    "database.internal.image.pullPolicy": "IfNotPresent",
    "database.internal.image.repository": "vmware/harbor-db",
    "database.internal.image.tag": "dev",
    "database.internal.password": "changeit",
    "database.internal.tolerations": [],
    "database.internal.volumes.data.accessMode": "ReadWriteOnce",
    "database.internal.volumes.data.size": "1Gi",
    "database.type": "internal",
    "email.from": "admin \u003csample_admin@mydomain.com\u003e",
    "email.host": "smtp.mydomain.com",
    "email.identity": "",
    "email.insecure": "false",
    "email.password": "password",
    "email.port": "25",
    "email.ssl": "false",
    "email.username": "sample_admin@mydomain.com",
    "externalDomain": "harbor.my.domain",
    "externalPort": 32700,
    "externalProtocol": "https",
    "harborAdminPassword": "Harbor12345",
    "harborImageTag": "dev",
    "ingress.annotations.ingress.kubernetes.io/proxy-body-size": "0",
    "ingress.annotations.ingress.kubernetes.io/ssl-redirect": "true",
    "ingress.annotations.nginx.ingress.kubernetes.io/proxy-body-size": "0",
    "ingress.annotations.nginx.ingress.kubernetes.io/ssl-redirect": "true",
    "ingress.enabled": true,
    "ingress.tls.secretName": "",
    "jobservice.image.pullPolicy": "IfNotPresent",
    "jobservice.image.repository": "vmware/harbor-jobservice",
    "jobservice.image.tag": "dev",
    "jobservice.maxWorkers": 50,
    "jobservice.secret": "not-a-secure-secret",
    "jobservice.tolerations": [],
    "ldap.baseDN": "",
    "ldap.filter": "(objectClass=person)",
    "ldap.scope": "2",
    "ldap.searchDN": "",
    "ldap.searchPassword": "",
    "ldap.timeout": "5",
    "ldap.uid": "uid",
    "ldap.url": "ldaps://ldapserver",
    "ldap.verifyCert": "True",
    "notary.enabled": true,
    "notary.server.image.pullPolicy": "IfNotPresent",
    "notary.server.image.repository": "vmware/notary-server-photon",
    "notary.server.image.tag": "dev",
    "notary.signer.caCrt": null,
    "notary.signer.env.NOTARY_SIGNER_DEFAULTALIAS": "defaultalias",
    "notary.signer.image.pullPolicy": "IfNotPresent",
    "notary.signer.image.repository": "vmware/notary-signer-photon",
    "notary.signer.image.tag": "dev",
    "notary.signer.tlsCrt": null,
    "notary.signer.tlsKey": null,
    "notary.tolerations": [],
    "persistence.enabled": true,
    "redis.cluster.enabled": false,
    "redis.external.databaseIndex": "0",
    "redis.external.enabled": false,
    "redis.external.host": "192.168.0.2",
    "redis.external.password": "changeit",
    "redis.external.port": "6379",
    "redis.external.usePassword": false,
    "redis.master.persistence.enabled": false,
    "redis.password": "changeit",
    "redis.usePassword": false,
    "registry.httpSecret": "not-a-secure-secret",
    "registry.image.pullPolicy": "IfNotPresent",
    "registry.image.repository": "vmware/registry-photon",
    "registry.image.tag": "dev",
    "registry.logLevel": "info",
    "registry.storage.azure.accountkey": "base64encodedaccountkey",
    "registry.storage.azure.accountname": "accountname",
    "registry.storage.azure.container": "containername",
    "registry.storage.filesystem.rootdirectory": "/var/lib/registry",
    "registry.storage.gcs.bucket": "bucketname",
    "registry.storage.oss.accesskeyid": "accesskeyid",
    "registry.storage.oss.accesskeysecret": "accesskeysecret",
    "registry.storage.oss.bucket": "bucketname",
    "registry.storage.oss.region": "regionname",
    "registry.storage.s3.bucket": "bucketname",
    "registry.storage.s3.region": "us-west-1",
    "registry.storage.swift.authurl": "https://storage.myprovider.com/v3/auth",
    "registry.storage.swift.container": "containername",
    "registry.storage.swift.password": "password",
    "registry.storage.swift.username": "username",
    "registry.storage.type": "filesystem",
    "registry.tolerations": [],
    "registry.volumes.data.accessMode": "ReadWriteOnce",
    "registry.volumes.data.size": "5Gi",
    "secretKey": "not-a-secure-key",
    "selfRegistration": "on",
    "ui.image.pullPolicy": "IfNotPresent",
    "ui.image.repository": "vmware/harbor-ui",
    "ui.image.tag": "dev",
    "ui.secret": "not-a-secure-secret",
    "ui.tolerations": []
  };
  yaml: string = "persistence:\n  enabled: true\nexternalProtocol: https\n# The FQDN for Harbor service\nexternalDomain: harbor.my.domain\n# The Port for Harbor service, leave empty if the service \n# is to be bound to port 80/443\nexternalPort: 32700\nharborAdminPassword: Harbor12345\nauthenticationMode: \"db_auth\"\nselfRegistration: \"on\"\nldap:\n  url: \"ldaps://ldapserver\"\n  searchDN: \"\"\n  searchPassword: \"\"\n  baseDN: \"\"\n  filter: \"(objectClass=person)\"\n  uid: \"uid\"\n  scope: \"2\"\n  timeout: \"5\"\n  verifyCert: \"True\"\nemail:\n  host: \"smtp.mydomain.com\"\n  port: \"25\"\n  username: \"sample_admin@mydomain.com\"\n  password: \"password\"\n  ssl: \"false\"\n  insecure: \"false\"\n  from: \"admin \u003csample_admin@mydomain.com\u003e\"\n  identity: \"\"\n\n# The secret key used for encryption. Must be a string of 16 chars.\nsecretKey: not-a-secure-key\n\n# These annotations allow the registry to work behind the nginx\n# ingress controller.\ningress:\n  enabled: true\n  annotations:\n    ingress.kubernetes.io/ssl-redirect: \"true\"\n    nginx.ingress.kubernetes.io/ssl-redirect: \"true\"\n    ingress.kubernetes.io/proxy-body-size: \"0\"\n    nginx.ingress.kubernetes.io/proxy-body-size: \"0\"\n  tls:\n    # Fill the secretName if you want to use the certificate of \n    # yourself when Harbor serves with HTTPS. A certificate will \n    # be generated automatically by the chart if leave it empty\n    secretName: \"\"\n\n# The tag for Harbor docker images.\nharborImageTag: \u0026harbor_image_tag dev\n\nadminserver:\n  image:\n    repository: vmware/harbor-adminserver\n    tag: *harbor_image_tag\n    pullPolicy: IfNotPresent\n  volumes:\n    config:\n      # storageClass: \"-\"\n      accessMode: ReadWriteOnce\n      size: 1Gi\n  # resources:\n  #  requests:\n  #    memory: 256Mi\n  #    cpu: 100m\n  nodeSelector: {}\n  tolerations: []\n  affinity: {}\n\njobservice:\n  image:\n    repository: vmware/harbor-jobservice\n    tag: *harbor_image_tag\n    pullPolicy: IfNotPresent\n  secret: not-a-secure-secret\n  maxWorkers: 50\n# resources:\n#   requests:\n#     memory: 256Mi\n#     cpu: 100m\n  nodeSelector: {}\n  tolerations: []\n  affinity: {}\n\nui:\n  image:\n    repository: vmware/harbor-ui\n    tag: *harbor_image_tag\n    pullPolicy: IfNotPresent\n  secret: not-a-secure-secret\n# resources:\n#  requests:\n#    memory: 256Mi\n#    cpu: 100m\n  nodeSelector: {}\n  tolerations: []\n  affinity: {}\n\n# TODO: change the style to be same with redis\ndatabase:\n  # if external database is used, set \"type\" to \"external\"\n  # and fill the connection informations in \"external\" section\n  type: internal\n  internal:\n    image:\n      repository: vmware/harbor-db\n      tag: *harbor_image_tag\n      pullPolicy: IfNotPresent\n    # the superuser password of database\n    password: \"changeit\"\n    volumes:\n      data:\n        # storageClass: \"-\"\n        accessMode: ReadWriteOnce\n        size: 1Gi\n    # resources:\n    #  requests:\n    #    memory: 256Mi\n    #    cpu: 100m\n    nodeSelector: {}\n    tolerations: []\n    affinity: {}\n  external:\n    host: \"192.168.0.1\"\n    port: \"5432\"\n    username: \"user\"\n    password: \"password\"\n    coreDatabase: \"registry\"\n    clairDatabase: \"clair\"\n    notaryServerDatabase: \"notary_server\"\n    notarySignerDatabase: \"notary_signer\"\n\nregistry:\n  image:\n    repository: vmware/registry-photon\n    tag: dev\n    pullPolicy: IfNotPresent\n  httpSecret: not-a-secure-secret\n  logLevel: info\n  storage:\n    # specify the type of storage: \"filesystem\", \"azure\", \"gcs\", \"s3\", \"swift\", \n    # \"oss\" and fill the information needed in the corresponding section\n    type: filesystem\n    filesystem:\n      rootdirectory: /var/lib/registry\n      #maxthreads: 100\n    azure:\n      accountname: accountname\n      accountkey: base64encodedaccountkey\n      container: containername\n      #realm: core.windows.net\n    gcs:\n      bucket: bucketname\n      # TODO: support the keyfile of gcs\n      #keyfile: /path/to/keyfile\n      #rootdirectory: /gcs/object/name/prefix\n      #chunksize: 5242880\n    s3:\n      region: us-west-1\n      bucket: bucketname\n      #accesskey: awsaccesskey\n      #secretkey: awssecretkey\n      #regionendpoint: http://myobjects.local\n      #encrypt: false\n      #keyid: mykeyid\n      #secure: true\n      #v4auth: true\n      #chunksize: 5242880\n      #rootdirectory: /s3/object/name/prefix\n      #storageclass: STANDARD\n    swift:\n      authurl: https://storage.myprovider.com/v3/auth\n      username: username\n      password: password\n      container: containername\n      #region: fr\n      #tenant: tenantname\n      #tenantid: tenantid\n      #domain: domainname\n      #domainid: domainid\n      #trustid: trustid\n      #insecureskipverify: false\n      #chunksize: 5M\n      #prefix:\n      #secretkey: secretkey\n      #accesskey: accesskey\n      #authversion: 3\n      #endpointtype: public\n      #tempurlcontainerkey: false\n      #tempurlmethods:\n    oss:\n      accesskeyid: accesskeyid\n      accesskeysecret: accesskeysecret\n      region: regionname\n      bucket: bucketname\n      #endpoint: endpoint\n      #internal: false\n      #encrypt: false\n      #secure: true\n      #chunksize: 10M\n      #rootdirectory: rootdirectory\n  ## Persist data to a persistent volume\n  volumes:\n    data:\n      # storageClass: \"-\"\n      accessMode: ReadWriteOnce\n      size: 5Gi\n  # resources:\n  #  requests:\n  #    memory: 256Mi\n  #    cpu: 100m\n  nodeSelector: {}\n  tolerations: []\n  affinity: {}\n\nchartmuseum:\n  enabled: true\n  image:\n    repository: vmware/chartmuseum-photon\n    tag: dev\n    pullPolicy: IfNotPresent\n  volumes:\n    data:\n      # storageClass: \"-\"\n      accessMode: ReadWriteOnce\n      size: 5Gi\n  # resources:\n  #  requests:\n  #    memory: 256Mi\n  #    cpu: 100m\n  nodeSelector: {}\n  tolerations: []\n  affinity: {}\n\nclair:\n  enabled: true\n  image:\n    repository: vmware/clair-photon\n    tag: dev\n    pullPolicy: IfNotPresent\n  volumes:\n    pgData:\n      # storageClass: \"-\"\n      accessMode: ReadWriteOnce\n      size: 1Gi\n  # resources:\n  #  requests:\n  #    memory: 256Mi\n  #    cpu: 100m\n  nodeSelector: {}\n  tolerations: []\n  affinity: {}\n\nredis:\n  # if external Redis is used, set \"external.enabled\" to \"true\"\n  # and fill the connection informations in \"external\" section.\n  # or the internal Redis will be used\n  usePassword: false\n  password: \"changeit\"\n  cluster:\n    enabled: false\n  master:\n    persistence:\n# TODO: There is a perm issue: Can't open the append-only file: Permission denied\n# TODO: Setting it to false is a temp workaround.  Will re-visit this problem.\n      enabled: false\n  external:\n    enabled: false\n    host: \"192.168.0.2\"\n    port: \"6379\"\n    databaseIndex: \"0\"\n    usePassword: false\n    password: \"changeit\"\n\nnotary:\n  enabled: true\n  server:\n    image:\n      repository: vmware/notary-server-photon\n      tag: dev\n      pullPolicy: IfNotPresent\n  signer:\n    image:\n      repository: vmware/notary-signer-photon\n      tag: dev\n      pullPolicy: IfNotPresent\n    env:\n      NOTARY_SIGNER_DEFAULTALIAS: defaultalias\n    # The TLS certificate for Notary Signer. Will auto generate them if unspecified here.\n    caCrt:\n    tlsCrt:\n    tlsKey:\n  nodeSelector: {}\n  tolerations: []\n  affinity: {}\n"
  ;

  // Default set to yaml file
  valueMode = false;
  valueHover = false;
  yamlHover = true;

  constructor(private errorHandler: ErrorHandler,
              private additionsService: AdditionsService) {
  }

  ngOnInit(): void {
    if (this.valuesLink && !this.valuesLink.absolute && this.valuesLink.href) {
      this.additionsService.getDetailByLink(this.valuesLink.href).subscribe(
        res => {
          this.values = res;
          this.yaml = JSON.stringify(res);
        }, error => {
          this.errorHandler.error(error);
        }
      );
    }
  }

  public get isValueMode() {
    return this.valueMode;
  }

  isHovering(view: string) {
    if (view === 'value') {
      return this.valueHover;
    } else {
      return this.yamlHover;
    }
  }

  showYamlFile(showYaml: boolean) {
    this.valueMode = !showYaml;
  }

  mouseEnter(mode: string) {
    if (mode === "value") {
      this.valueHover = true;
    } else {
      this.yamlHover = true;
    }
  }

  mouseLeave(mode: string) {
    if (mode === "value") {
      this.valueHover = false;
    } else {
      this.yamlHover = false;
    }
  }
}
