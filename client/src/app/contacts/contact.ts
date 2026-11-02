// Модель контакта. _id не заполняется на клиенте — его создаёт MongoDB (ObjectId)
export interface Contact {
  _id?: string;
  name: string;
  email: string;
  phone: {
    mobile: string;
    work: string;
  };
}
