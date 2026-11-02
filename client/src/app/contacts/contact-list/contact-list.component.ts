// Список контактов: загружает данные при старте и хранит локальную копию,
// чтобы после создания/изменения/удаления не запрашивать весь список заново
import { Component, OnInit, inject } from '@angular/core';
import { Contact } from '../contact';
import { ContactService } from '../contact.service';
import { ContactDetailsComponent } from '../contact-details/contact-details.component';

@Component({
  selector: 'contact-list',
  imports: [ContactDetailsComponent],
  templateUrl: './contact-list.component.html',
})
export class ContactListComponent implements OnInit {
  private contactService = inject(ContactService);
  contacts: Contact[] = [];
  selectedContact: Contact | null = null;
  error = '';

  async ngOnInit() {
    try {
      this.contacts = await this.contactService.getContacts();
    } catch {
      this.error = 'Не удалось загрузить контакты';
    }
  }

  private indexOf(id?: string) {
    return this.contacts.findIndex((c) => c._id === id);
  }

  selectContact(contact: Contact | null) {
    this.selectedContact = contact;
  }

  // Новый (ещё не сохранённый) контакт — без _id
  createNewContact() {
    this.selectContact({ name: '', email: '', phone: { mobile: '', work: '' } });
  }

  // Обработчики передаются в компонент деталей и вызываются после ответа сервера
  addContact = (contact: Contact) => {
    this.contacts.push(contact);
    this.selectContact(contact);
  };

  updateContact = (contact: Contact) => {
    const i = this.indexOf(contact._id);
    if (i !== -1) {
      this.contacts[i] = contact;
      this.selectContact(contact);
    }
  };

  deleteContact = (id: string) => {
    const i = this.indexOf(id);
    if (i !== -1) {
      this.contacts.splice(i, 1);
      this.selectContact(null);
    }
  };
}
