import {
  Component,
  OnInit,
  Input,
  ViewChild,
  Output,
  EventEmitter,
} from "@angular/core";
import { Webhook } from "../webhook";
import { AddWebhookFormComponent } from "../add-webhook-form/add-webhook-form.component";

@Component({
  selector: 'add-webhook',
  templateUrl: './add-webhook.component.html',
  styleUrls: ['./add-webhook.component.scss']
})
export class AddWebhookComponent implements OnInit {
  isOpen: boolean = false;
  closable: boolean = true;
  staticBackdrop: boolean = true;

  @Input() projectId: number;
  @Input() webhook: Webhook;
  @Output() modify = new EventEmitter<boolean>();
  @ViewChild(AddWebhookFormComponent)
  addWebhookFormComponent: AddWebhookFormComponent;


  constructor() { }

  ngOnInit() {
  }

  openAddWebhookModal() {
    this.isOpen = true;
  }

  onCancel() {
    this.isOpen = false;
  }

  closeModal(isModified: boolean): void {
    if (isModified) {
      this.modify.emit(true);
    }
    this.isOpen = false;
  }

}
