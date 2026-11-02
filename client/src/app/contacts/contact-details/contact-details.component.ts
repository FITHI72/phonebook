// Карточка контакта: форма и кнопки создания/обновления/удаления.
// Сначала отправляет запрос на сервер, затем сообщает списку о результате.
import { Component, Input, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Contact } from '../contact';
import { ContactService } from '../contact.service';

@Component({
  selector: 'contact-details',
  imports: [FormsModule],
  templateUrl: './contact-details.component.html',
})
export class ContactDetailsComponent {
  private contactService = inject(ContactService);
  @Input() contact: Contact | null = null;
  @Input() createHandler!: (c: Contact) => void;
  @Input() updateHandler!: (c: Contact) => void;
  @Input() deleteHandler!: (id: string) => void;
  message = '';

  async createContact(contact: Contact) {
    try {
      const created = await this.contactService.createContact(contact);
      this.message = '';
      this.createHandler(created);
    } catch (e: any) {
      this.message = e?.error?.error || 'Ошибка сервера';
    }
  }

  async updateContact(contact: Contact) {
    try {
      const updated = await this.contactService.updateContact(contact);
      this.message = '';
      this.updateHandler(updated);
    } catch (e: any) {
      this.message = e?.error?.error || 'Ошибка сервера';
    }
  }

  async deleteContact(id: string) {
    try {
      const res = await this.contactService.deleteContact(id);
      this.message = '';
      this.deleteHandler(res._id);
    } catch (e: any) {
      this.message = e?.error?.error || 'Ошибка сервера';
    }
  }
}
