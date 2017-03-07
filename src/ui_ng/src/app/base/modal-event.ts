import { modalEvents } from './modal-events.const'

//Define a object to store the modal event
export class ModalEvent {
    modalName: modalEvents;
    modalFlag: boolean; //true for open, false for close
}