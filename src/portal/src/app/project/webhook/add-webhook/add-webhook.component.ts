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
  isEdit: boolean;
  isOpen: boolean = false;
  closable: boolean = false;
  staticBackdrop: boolean = true;

  @Input() projectId: number;
   webhook: Webhook;
  @Input()
  metadata: any;
  @ViewChild(AddWebhookFormComponent, { static: false })
  addWebhookFormComponent: AddWebhookFormComponent;
  @Output() notify = new EventEmitter<Webhook>();

  constructor() { }

  ngOnInit() {
  }

  openAddWebhookModal() {
    this.isOpen = true;
  }

  onCancel() {
    this.isOpen = false;
  }
  notifySuccess() {
    this.isOpen = false;
    this.notify.emit();
  }
  closeModal() {
    this.isOpen = false;
  }
}
