import { Component, Input, OnInit } from "@angular/core";
import { ErrorHandler } from "../../../../utils/error-handler";
import { AdditionsService } from "../additions.service";
import { ArtifactBuildHistory } from "../models";

@Component({
  selector: "hbr-artifact-build-history",
  templateUrl: "./build-history.component.html",
  styleUrls: ["./build-history.component.scss"],
})
export class BuildHistoryComponent implements OnInit {
  @Input()
  buildHistoryLink: string;
  historyList: ArtifactBuildHistory[] = [];
  loading: Boolean = false;
  // todo
  demo = "{\"architecture\":\"amd64\",\"config\":{\"Hostname\":\"\",\"Domainname\":\"\",\"User\":\"\"," +
    "\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\"" +
    ":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:" +
    "/sbin:/bin\"],\"Cmd\":[\"/bin/bash\"],\"ArgsEscaped\":true,\"Image\":\"sha256:cfd729f746ff" +
    "94efd76b0b4a4baf04635641a8b47ace75d21a9ea8689e466929\",\"Volumes\":null,\"WorkingDir\":\"\"," +
    "\"Entrypoint\":null,\"OnBuild\":null,\"Labels\":{\"build-date\":\"20191025\",\"name\":\"Photon " +
    "OS 2.0 Base Image\",\"vendor\":\"VMware\"}},\"container\":\"eb3b5f093ce9e32d76083e55e9a6d7d1811af" +
    "1ea57065250d8f2caeb865689e3\",\"container_config\":{\"Hostname\":\"\",\"Domainname\":\"\"," +
    "\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\"" +
    ":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local" +
    "/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"tdnf install -y cronie " +
    "rsyslog logrotate shadow tar gzip sudo \\u003e\\u003e /dev/null    \\u0026\\u0026 mkdir /var/" +
    "spool/rsyslog     \\u0026\\u0026 groupadd -r -g 10000 syslog \\u0026\\u0026 useradd --no-log-" +
    "init -r -g 10000 -u 10000 syslog     \\u0026\\u0026 tdnf clean all\"],\"Image\":\"sha256:cfd7" +
    "29f746ff94efd76b0b4a4baf04635641a8b47ace75d21a9ea8689e466929\",\"Volumes\":null,\"Work" +
    "ingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":null,\"Labels\":{\"build-date\":\"201910" +
    "25\",\"name\":\"Photon OS 2.0 Base Image\",\"vendor\":\"VMware\"}},\"created\":\"2019-1" +
    "1-11T08:37:49.683105872Z\",\"docker_version\":\"19.03.2\",\"history\":[{\"created\":\"201" +
    "9-10-28T21:26:10.234058084Z\",\"created_by\":\"/bin/sh -c #(nop) ADD file:383fe94c8068ad68" +
    "dc54778061b329b5d136d9d9c8911f2605271e2f82eb7e1a in / \"},{\"created\":\"2019-10-28T21:26:" +
    "10.399504887Z\",\"created_by\":\"/bin/sh -c #(nop)  LABEL name=Photon OS 2.0 Base Image vendo" +
    "r=VMware build-date=20191025\",\"empty_layer\":true},{\"created\":\"2019-10-28T21:26:10.562" +
    "853047Z\",\"created_by\":\"/bin/sh -c #(nop)  CMD [\\\"/bin/bash\\\"]\",\"empty_layer\":tru" +
    "e},{\"created\":\"2019-11-11T08:37:49.683105872Z\",\"created_by\":\"/bin/sh -c tdnf install " +
    "-y cronie rsyslog logrotate shadow tar gzip sudo \\u003e\\u003e /dev/null    \\u0026\\u002" +
    "6 mkdir /var/spool/rsyslog     \\u0026\\u0026 groupadd -r -g 10000 syslog \\u0026\\u0026 u" +
    "seradd --no-log-init -r -g 10000 -u 10000 syslog     \\u0026\\u0026 tdnf cl" +
    "ean all\"}],\"os\":\"linux\",\"rootfs\":{\"type\":\"layers\",\"diff_ids\":[\"sha" +
    "256:47a4bb1cfbc75e1073f8b6fc0806588bbde1142cae1ff0be70d4a111aaa05b0d\",\"sha256:69e4324" +
    "2ff643cec50cac17983cda8dd22d16707daa8b6652e17aae00362d501\"]}}";

  constructor(
    private errorHandler: ErrorHandler,
    private additionsService: AdditionsService
  ) {
  }

  ngOnInit(): void {
    JSON.parse(this.demo).history.forEach((ele: any) => {
      const history: ArtifactBuildHistory = new ArtifactBuildHistory();
      history.createdTime = ele.created;
      if (ele.created_by !== undefined) {
        history.createdBy = ele.created_by
          .replace("/bin/sh -c #(nop)", "")
          .trimLeft()
          .replace("/bin/sh -c", "RUN");
      } else {
        history.createdBy = ele.comment;
      }
      this.historyList.push(history);
    });
    if (this.buildHistoryLink) {
      this.additionsService.getDetailByLink(this.buildHistoryLink).subscribe(
        res => {
          if (res && res.length) {
            res.forEach((ele: any) => {
              const history: ArtifactBuildHistory = new ArtifactBuildHistory();
              history.createdTime = ele.created;
              if (ele.created_by !== undefined) {
                history.createdBy = ele.created_by
                  .replace("/bin/sh -c #(nop)", "")
                  .trimLeft()
                  .replace("/bin/sh -c", "RUN");
              } else {
                history.createdBy = ele.comment;
              }
              this.historyList.push(history);
            });
          }
        }, error => {
          this.errorHandler.error(error);
        }
      );
    }
  }
}
