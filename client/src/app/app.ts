import { Component } from '@angular/core';
import { ContactListComponent } from './contacts/contact-list/contact-list.component';

@Component({
  selector: 'app-root',
  imports: [ContactListComponent],
  templateUrl: './app.html',
})
export class App {}
