import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from "@angular/core";
import { AdditionsService } from "../additions.service";
import { of } from "rxjs";
import { SummaryComponent } from "./summary.component";
import { AdditionLink } from "../../../../../../../ng-swagger-gen/models/addition-link";
import { IServiceConfig, SERVICE_CONFIG } from "../../../../../../lib/entities/service.config";
import { ErrorHandler } from "../../../../../../lib/utils/error-handler";
import { ProjectModule } from "../../../../project.module";
import { CURRENT_BASE_HREF } from "../../../../../../lib/utils/utils";

describe('SummaryComponent', () => {
  let component: SummaryComponent;
  let fixture: ComponentFixture<SummaryComponent>;
  const mockedLink: AdditionLink = {
    absolute: false,
    href: '/test'
  };
  const readme: string = "# Helm Chart for Harbor\n\n## Introduction\n\nThis [Helm](https://github.com/" +
    "kubernetes/helm) chart installs [Harbor](http://vmware.github.io/harbor/) in a Kubernetes " +
    "cluster. Currently this chart supports Harbor v1.4.0 release. Welcome to [contribute](CONTR" +
    "IBUTING.md) to Helm Chart for Harbor.\n\n## Prerequisites\n\n- Kubernetes cluster 1.8+ with " +
    "Beta APIs enabled\n- Kubernetes Ingress Controller is enabled\n- kubectl CLI 1.8+\n- Helm CLI" +
    " 2.8.0+\n\n## Known Issues\n\n- This chart doesn't work with Kubernetes security update release" +
    " 1.8.9+ and 1.9.4+. Refer to [issue 4496](https://github.com/vmware/harbor/issues/4496).\n\n## " +
    "Setup a Kubernetes cluster\n\nYou can use any tools to setup a K8s cluster.\nIn this guide," +
    " we use [minikube](https://github.com/kubernetes/minikube) 0.25.0 to setup a K8s cluster as " +
    "the dev/test env.\n```bash\n# Start minikube\nminikube start --vm-driver=none\n# Enable Ingress" +
    " Controller\nminikube addons enable ingress\n```\n## Installing the Chart\n\nFirst install" +
    " [Helm CLI](https://github.com/kubernetes/helm#install), then initialize Helm.\n```bash\nhelm" +
    " init\n```\nDownload Harbor helm chart code.\n```bash\ngit clone https://github.com/vmware/" +
    "harbor\ncd harbor/contrib/helm/harbor\n```\nDownload external dependent charts required by" +
    " Harbor chart.\n```bash\nhelm dependency update\n```\n### Secure Registry Mode\n\nBy default " +
    "this chart will generate a root CA and SSL certificate for your Harbor.\nYou can also use your" +
    " own CA signed certificate:\n\nopen values.yaml, set the value of 'externalDomain' to" +
    " your Harbor FQDN, and\nset value of 'tlsCrt', 'tlsKey', 'caCrt'. The common name of the " +
    "certificate must match your Harbor FQDN.\n\nInstall the Harbor helm chart with a release " +
    "name `my-release`:\n```bash\nhelm install . --debug --name my-release --set externalDomain" +
    "=harbor.my.domain\n```\n**Make sure** `harbor.my.domain` resolves to the K8s Ingress Contr" +
    "oller IP on the machines where you run docker or access Harbor UI.\nYou can add `harbor.my.domain`" +
    " and IP mapping in the DNS server, or in /etc/hosts, or use the FQDN `harbor.\u003cIP\u003e.xip." +
    "io`.\n\nFollow the `NOTES` section in the command output to get Harbor admin password and **add " +
    "Harbor root CA into docker trusted certificates**.\n\nIf you are using an external service like " +
    "[cert-manager](https://github.com/jetstack/cert-manager) for generating the TLS certificates,\nyou" +
    "will want to disable the certificate generation by helm by setting the value `generateCertificates` " +
    "to _false_. Then the ingress' annotations will be scanned\nby _cert-manager_ and the appropriate " +
    "secret will get created and updated by the service.\n\nIf using acme's certificates, do not forget to " +
    "add the following annotation to\nyour ingress.\n\n```yaml\ningress:\n  annotations:\n    kubernetes.io/" +
    "tls-acme: \"true\"\n```\n\nThe command deploys Harbor on the Kubernetes cluster in the default " +
    "configuration.\nThe [configuration](#configuration) section lists the parameters that can be configured" +
    " in values.yaml or via '--set' params during installation.\n\n\u003e **Tip**: List all releases using" +
    " `helm list`\n\n\n### Insecure Registry Mode\n\nIf setting Harbor Registry as insecure-registries for " +
    "docker,\nyou don't need to generate Root CA and SSL certificate for the Harbor ingress controller.\n\nInstal" +
    "l the Harbor helm chart with a release name `my-release`:\n```bash\nhelm install . --debug --name my-release" +
    " --set externalDomain=harbor.my.domain,insecureRegistry=true\n```\n**Make sure** `harbor.my.domain` resolves" +
    " to the K8s Ingress Controller IP on the machines where you run docker or access Harbor UI.\nYou can add" +
    " `harbor.my.domain` and IP mapping in the DNS server, or in /etc/hosts, or use the FQDN " +
    "`harbor.\u003cIP\u003e.xip.io`.\n\nThen add `\"insecure-registries\": [\"harbor.my.domain\"]`" +
    " in the docker daemon config file and restart docker service.\n\n## Uninstalling the Chart\n\nTo " +
    "uninstall/delete the `my-release` deployment:\n\n```bash\nhelm delete my-release\n```\n\nThe command " +
    "removes all the Kubernetes components associated with the chart and deletes the release.\n\n## " +
    "Configuration\n\nThe following tables lists the configurable parameters of the Harbor chart and the " +
    "default values.\n\n| Parameter                  | Description                        " +
    "| Default                 |\n| -----------------------    | ---------------------------------- | ----" +
    "------------------- |\n| **Harbor** |\n| `harborImageTag`     | The tag for Harbor docker images | " +
    "`v1.4.0` |\n| `externalDomain`       | Harbor will run on (https://`externalDomain`/). Recommend using" +
    " K8s Ingress Controller FQDN as `externalDomain`, or make sure this FQDN resolves to the K8s Ingress" +
    " Controller IP. | `harbor.my.domain` |\n| `insecureRegistry`     | If set to true, you don't need to" +
    " set tlsCrt/tlsKey/caCrt, but must add Harbor FQDN as insecure-registries for your docker client. " +
    "| `false` |\n| `generateCertificates`  | Set to false if TLS certificate will be managed by an external " +
    "service | `true` |\n| `tlsCrt`               | TLS certificate to use for Harbor's https endpoint. Its" +
    " CN must match `externalDomain`. | auto-generated |\n| `tlsKey`               | TLS key to use for " +
    "Harbor's https endpoint | auto-generated |\n| `caCrt`                | CA Cert for self signed TLS cert" +
    " | auto-generated |\n| `persistence.enabled` | enable persistent data storage | `false` |\n| `secretKey` " +
    "| The secret key used for encryption. Must be a string of 16 chars. | " +
    "`not-a-secure-key` |\n| **Adminserver** |\n| `adminserver.image.repository`" +
    " | Repository for adminserver image | `vmware/harbor-adminserver` |\n| `adminserver.image.tag`" +
    " | Tag for adminserver image | `v1.4.0` |\n| `adminserver.image.pullPolicy` | " +
    "Pull Policy for adminserver image | `IfNotPresent` |\n| `adminserver.emailHost` |" +
    " email server | `smtp.mydomain.com` |\n| `adminserver.emailPort` | email port | `25` |\n| " +
    "`adminserver.emailUser` | email username | `sample_admin@mydomain.com` |\n| `adminserver.emailSsl` " +
    "| email uses SSL? | `false` |\n| `adminserver.emailFrom` | send email from address | `admin \u003csample_admin@" +
    "mydomain.com\u003e` |\n| `adminserver.emailIdentity` | | \"\" |\n| `adminserver.key` | adminsever key | " +
    "`not-a-secure-key` |\n| `adminserver.emailPwd` | password for email | `not-a-secure-password` |\n| `adminserver." +
    "adminPassword` | password for admin user | `Harbor12345` |\n| `adminserver.authenticationMode` | authentication" +
    " mode for Harbor ( `db_auth` for local database, `ldap_auth` for LDAP, etc...) [Docs](https://github.com/vmware/" +
    "harbor/blob/master/docs/user_guide.md#user-account) | `db_auth` |\n| `adminserver.selfRegistration` | Allows users" +
    " to register by themselves, otherwise only administrators can add users | `on` |\n| `adminserver.ldap.url` | LDAP" +
    " server URL for `ldap_auth` authentication | `ldaps://ldapserver` |\n| `adminserver.ldap.searchDN` |" +
    " LDAP Search DN | `` |\n| `adminserver.ldap.baseDN` | LDAP Base DN | `` |\n| `adminserver.ldap.filter` | LDAP Filter " +
    "| `(objectClass=person)` |\n| `adminserver.ldap.uid` | LDAP UID | `uid` |\n| `adminserver.ldap.scope` | LDAP Scope" +
    " | `2` |\n| `adminserver.ldap.timeout` | LDAP Timeout | `5` |\n| `adminserver.ldap.verifyCert` | LDAP Verify " +
    "HTTPS Certificate | `True` |\n| `adminserver.resources` | [resources](https://kubernetes.io/docs/concepts" +
    "/configuration/manage-compute-resources-container/) to allocate for container   | undefined |\n| " +
    "`adminserver.volumes` | used to create PVCs if persistence is enabled (see instructions in values.yaml) | " +
    "see values.yaml |\n| `adminserver.nodeSelector` | Node labels for pod assignment | `{}` |\n| `adminserver." +
    "tolerations` | Tolerations for pod assignment | `[]` |\n| `adminserver.affinity` | Node/Pod affinities " +
    "| `{}` |\n| **Jobservice** |\n| `jobservice.image.repository` | Repository for jobservice image | `vmware" +
    "/harbor-jobservice` |\n| `jobservice.image.tag` | Tag for jobservice image | `v1.4.0` |\n| `jobservice." +
    "image.pullPolicy` | Pull Policy for jobservice image | `IfNotPresent` |\n| `jobservice.key` | jobservice" +
    "key | `not-a-secure-key` |\n| `jobservice.secret` | jobservice secret | `not-a-secure-secret` |\n| " +
    "`jobservice.resources` | [resources](https://kubernetes.io/docs/concepts/configuration/" +
    "manage-compute-resources-container/) to allocate for container   | undefined |\n| `jobservice.nodeSelector` " +
    "| Node labels for pod assignment | `{}` |\n| `jobservice.tolerations` | Tolerations for pod assignment |" +
    " `[]` |\n| `jobservice.affinity` | Node/Pod affinities | `{}` |\n| **UI** |\n| `ui.image.repository` | " +
    "epository for ui image | `vmware/harbor-ui` |\n| `ui.image.tag` | Tag for ui image | `v1.4.0` |\n| `ui." +
    "image.pullPolicy` | Pull Policy for ui image | `IfNotPresent` |\n| `ui.key` | ui key | `not-a-secure-key" +
    "` |\n| `ui.secret` | ui secret | `not-a-secure-secret` |\n| `ui.privateKeyPem` | ui private key | see " +
    "values.yaml |\n| `ui.resources` | [resources](https://kubernetes.io/docs/concepts/configuration/manage-" +
    "compute-resources-container/) to allocate for container  " +
    " | undefined |\n| `ui.nodeSelector` | Node labels for pod assignment " +
    "| `{}` |\n| `ui.tolerations` | Tolerations for pod assignment | `[]` |\n| `ui.affinity` | Node/Pod affinities" +
    " | `{}` |\n| **MySQL** |\n| `mysql.image.repository` | Repository for mysql image | `vmware/harbor-mysql` " +
    "|\n| `mysql.image.tag` | Tag for mysql image | `v1.4.0` |\n| `mysql.image.pullPolicy` | Pull Policy for mysql " +
    "image | `IfNotPresent` |\n| `mysql.host` | MySQL Server | `~` |\n| `mysql.port` | MySQL Port | `3306` |\n| " +
    "`mysql.user` | MySQL Username | `root` |\n| `mysql.pass` | MySQL Password | `registry` |\n| " +
    "`mysql.database` | MySQL Database | `registry` |\n| `mysql.resources` | [resources](https://kubernetes.io/" +
    "docs/concepts/configuration/manage-compute-resources-container/) to allocate for container   | undefined |\n| " +
    "`mysql.volumes` | used to create PVCs if persistence is enabled (see instructions in values.yaml) | " +
    "see values.yaml |\n| `mysql.nodeSelector` | Node labels for pod assignment | `{}` |\n| `mysql.tolerations` " +
    "| Tolerations for pod assignment | `[]` |\n| `mysql.affinity` | Node/Pod affinities" +
    " | `{}` |\n| **Registry** |\n| `registry.image.repository` | Repository for registry image | `" +
    "vmware/registry-photon` |\n| `registry.image.tag` | Tag for registry image | `v2.6.2-v1.4.0` |\n| " +
    "`registry.image.pullPolicy` | Pull Policy for registry image | `IfNotPresent` |\n| `registry.rootCrt` | " +
    "registry root cert " +
    "| see values.yaml |\n| `registry.httpSecret` | registry secret | `not-a-secure-secret` |\n| `registry.resources` " +
    "| [resources](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) to allocate" +
    " for container   | undefined |\n| `registry.volumes` | used to create PVCs if persistence is enabled (see " +
    "instructions in values.yaml) | see values.yaml |\n| `registry.nodeSelector` | Node labels for pod assignment " +
    "| `{}` |\n| `registry.tolerations` | Tolerations for pod assignment | `[]` |\n| `registry.affinity` | " +
    "Node/Pod affinities | `{}` |\n| **Clair** |\n| `clair.enabled` | Enable Clair? | `true` |\n| " +
    "`clair.image.repository` | Repository for clair image | `vmware/clair-photon` |\n| `clair.image.tag` |" +
    " Tag for clair image | `v2.0.1-v1.4.0`\n| `clair.resources` | [resources](https://kubernetes.io/docs/concepts/" +
    "configuration/manage-compute-resources-container/) to allocate for container   | undefined\n| `clair.nodeSelector" +
    "` | Node labels for pod assignment | `{}` |\n| `clair.tolerations` | Tolerations for pod assignment | `[]` |\n| " +
    "`clair.affinity` | Node/Pod affinities | `{}` |\n| `postgresql` | Overrides for postgresql chart [values.yaml](https" +
    "://github.com/kubernetes/charts/blob/f2938a46e3ae8e2512ede1142465004094c3c333/stable/postgresql/values.yaml) | " +
    "see values.yaml\n| **Notary** |\n| `notary.enabled` | Enable Notary? | `true` |\n| `notary.server.image.repository`" +
    " | Repository for notary server image | `vmware/notary-server-photon` |\n| `notary.server.image.tag` | Tag for " +
    "notary server image | `v0.5.1-v1.4.0`\n| `notary.signer.image.repository` | Repository for notary signer image |" +
    " `vmware/notary-signer-photon` |\n| `notary.signer.image.tag` | Tag for notary signer image | `v0.5.1-v1.4.0`\n|" +
    " `notary.db.image.repository` | Repository for notary database image | `vmware/mariadb-photon` |\n|" +
    "`notary.db.image.tag` | Tag for notary database image | `v1.4.0`\n| `notary.db.password` | The password of users " +
    "for notary database | Specify your own password |\n| `notary.nodeSelector` | Node labels for pod assignment " +
    "| `{}` |\n| `notary.tolerations` | Tolerations for pod assignment | `[]` |\n| `notary.affinity` | " +
    "Node/Pod affinities | `{}` |\n| **Ingress** |\n| `ingress.enabled` | Enable ingress objects. | `true` " +
    "|\n\nSpecify each parameter using the `--set key=value[,key=value]` argument to `helm install`. " +
    "For example:\n\n```bash\nhelm install . --name my-release --set externalDomain=" +
    "harbor.\u003cIP\u003e.xip.io\n```\n\nAlternatively," +
    " a YAML file that specifies the values for the parameters can be provided while installing the chart. For " +
    "example,\n\n```bash\nhelm install . --name my-release -f /path/to/values.yaml\n```\n\n\u003e **Tip**: " +
    "You can use the default [values.yaml](values.yaml)\n\n## Persistence\n\nHarbor stores the data and " +
    "configurations in emptyDir volumes. You can change the values.yaml to enable persistence and use a " +
    "PersistentVolumeClaim instead.\n\n\u003e *\"An emptyDir volume is first created when a Pod is " +
    "assigned to a Node, and exists as long as that Pod is running on that node. When a Pod is removed " +
    "from a node for any reason, the data in the emptyDir is deleted forever.\"*\n";

  const fakedAdditionsService = {
    getDetailByLink() {
      return of(readme);
    }
  };
  const config: IServiceConfig = {
    repositoryBaseEndpoint: CURRENT_BASE_HREF + "/repositories/testing"
  };
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
       ProjectModule
      ],
      providers: [
        ErrorHandler,
        { provide: AdditionsService, useValue: fakedAdditionsService },
        { provide: SERVICE_CONFIG, useValue: config },
      ],
      schemas: [
        NO_ERRORS_SCHEMA
      ]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(SummaryComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
  it('should get readme  and render', async () => {
    component.summaryLink = mockedLink;
    component.ngOnInit();
    fixture.detectChanges();
    await fixture.whenStable();
    const tables = fixture.nativeElement.getElementsByTagName('table');
    expect(tables.length).toEqual(1);
  });
});
