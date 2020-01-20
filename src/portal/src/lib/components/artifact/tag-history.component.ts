import { Component, Input, Output, EventEmitter, OnInit } from "@angular/core";
import { TagService, Manifest } from "../../services";
import { ErrorHandler } from "../../utils/error-handler";

@Component({
  selector: "hbr-tag-history",
  templateUrl: "./tag-history.component.html",
  styleUrls: ["./tag-history.component.scss"],

  providers: []
})
export class TagHistoryComponent implements OnInit {
  @Input()
  tagId: string;
  @Input()
  repositoryId: string;

  @Output()
  backEvt: EventEmitter<any> = new EventEmitter<any>();

  config: any = {};
  history: Object[] = [];
  loading: Boolean = false;

  constructor(
    private tagService: TagService,
    private errorHandler: ErrorHandler
  ) {}

  ngOnInit(): void {
    if (this.repositoryId && this.tagId) {
      this.retrieve(this.repositoryId, this.tagId);
    }
  }

  retrieve(repositoryId: string, tagId: string) {
    this.loading = true;
      this.tagService.getManifest(this.repositoryId, this.tagId)
      .subscribe(data => {
        this.config = JSON.parse(data.config);
        this.config.history.forEach((ele: any) => {
          if (ele.created_by !== undefined) {
            ele.created_by = ele.created_by
              .replace("/bin/sh -c #(nop)", "")
              .trimLeft()
              .replace("/bin/sh -c", "RUN");
          } else {
            ele.created_by = ele.comment;
          }
          this.history.push(ele);
        });
        this.loading = false;
      }, error => {
        this.errorHandler.error(error);
        this.loading = false;
      });
  }

  onBack(): void {
    this.backEvt.emit(this.tagId);
  }
}
