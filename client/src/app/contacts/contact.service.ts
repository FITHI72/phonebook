// Сервис — промежуточный слой между компонентами и REST API сервера
import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';
import { Contact } from './contact';

@Injectable({ providedIn: 'root' })
export class ContactService {
  private http = inject(HttpClient);
  private apiUrl = '/api/contacts'; // базовый путь API

  // GET /api/contacts — все контакты
  getContacts(): Promise<Contact[]> {
    return firstValueFrom(this.http.get<Contact[]>(this.apiUrl));
  }

  // POST /api/contacts — создать контакт
  createContact(contact: Contact): Promise<Contact> {
    return firstValueFrom(this.http.post<Contact>(this.apiUrl, contact));
  }

  // PUT /api/contacts/:id — обновить контакт
  updateContact(contact: Contact): Promise<Contact> {
    return firstValueFrom(this.http.put<Contact>(`${this.apiUrl}/${contact._id}`, contact));
  }

  // DELETE /api/contacts/:id — удалить контакт, сервер возвращает его id
  deleteContact(id: string): Promise<{ _id: string }> {
    return firstValueFrom(this.http.delete<{ _id: string }>(`${this.apiUrl}/${id}`));
  }
}
