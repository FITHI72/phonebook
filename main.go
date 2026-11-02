package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// phone вложенный объект с телефонами
type Phone struct {
	Mobile string `json:"mobile" bson:"mobile"`
	Work   string `json:"work" bson:"work"`
}

// contact  документ коллекции contacts.
// теги json  имена полей в HTTP-ответах, bson  имена полей в MongoDB.
type Contact struct {
	ID        bson.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name      string        `json:"name" bson:"name"`
	Email     string        `json:"email" bson:"email"`
	Phone     Phone         `json:"phone" bson:"phone"`
	CreatedAt *time.Time    `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt *time.Time    `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

// contactInput то что принимаем от клиента
type contactInput struct {
	Name  *string `json:"name"`
	Email string  `json:"email"`
	Phone Phone   `json:"phone"`
}

var contacts *mongo.Collection // коллекция contacts, создаётся один раз при старте

// getenv возвращает переменную окружения или значение по умолчанию
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// writeJSON отправляет ответ с кодом и телом
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

// handleError общий обработчик ошибок
func handleError(w http.ResponseWriter, reason, message string, code int) {
	log.Println("Ошибка:", reason)
	writeJSON(w, code, map[string]string{"error": message})
}

// чек тела запроса и проверка что имя указано
func readContact(r *http.Request) (Contact, string) {
	var in contactInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return Contact{}, "Тело запроса должно быть JSON-объектом"
	}
	if in.Name == nil || strings.TrimSpace(*in.Name) == "" {
		return Contact{}, "Не указано имя контакта (поле name)"
	}
	return Contact{Name: strings.TrimSpace(*in.Name), Email: in.Email, Phone: in.Phone}, ""
}

// parseID превращает строку из адреса в ObjectID
func parseID(w http.ResponseWriter, r *http.Request) (bson.ObjectID, bool) {
	id, err := bson.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		handleError(w, "Неверный id", "Неверный формат id", http.StatusBadRequest)
		return id, false
	}
	return id, true
}

// GET /api/contacts  все контакты по алфавиту
func listContacts(w http.ResponseWriter, r *http.Request) {
	cur, err := contacts.Find(r.Context(), bson.D{}, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		handleError(w, err.Error(), "Не удалось получить контакты", http.StatusInternalServerError)
		return
	}
	list := []Contact{} // пустой срез, а не nil  чтобы в JSON был [], а не null
	if err := cur.All(r.Context(), &list); err != nil {
		handleError(w, err.Error(), "Не удалось получить контакты", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// POST /api/contacts  создать контакт
func createContact(w http.ResponseWriter, r *http.Request) {
	c, problem := readContact(r)
	if problem != "" {
		handleError(w, "Неверные данные", problem, http.StatusBadRequest)
		return
	}
	now := time.Now().UTC().Truncate(time.Millisecond) // точность как в MongoDB  миллисекунды
	c.CreatedAt = &now
	res, err := contacts.InsertOne(r.Context(), c)
	if err != nil {
		handleError(w, err.Error(), "Не удалось создать контакт", http.StatusInternalServerError)
		return
	}
	c.ID = res.InsertedID.(bson.ObjectID) // _id, который сгенерировала MongoDB
	writeJSON(w, http.StatusCreated, c)
}

// DELETE /api/contacts  удалить все контакты
func deleteAll(w http.ResponseWriter, r *http.Request) {
	res, err := contacts.DeleteMany(r.Context(), bson.D{})
	if err != nil {
		handleError(w, err.Error(), "Не удалось удалить контакты", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"deletedCount": res.DeletedCount})
}

// GET /api/contacts/{id}  один контакт
func getContact(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var c Contact
	err := contacts.FindOne(r.Context(), bson.D{{Key: "_id", Value: id}}).Decode(&c)
	if errors.Is(err, mongo.ErrNoDocuments) {
		handleError(w, "Нет контакта", "Контакт не найден", http.StatusNotFound)
		return
	}
	if err != nil {
		handleError(w, err.Error(), "Не удалось получить контакт", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// PUT /api/contacts/{id}  обновить контакт
func updateContact(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	c, problem := readContact(r)
	if problem != "" {
		handleError(w, "Неверные данные", problem, http.StatusBadRequest)
		return
	}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "name", Value: c.Name},
		{Key: "email", Value: c.Email},
		{Key: "phone", Value: c.Phone},
		{Key: "updatedAt", Value: time.Now().UTC()},
	}}}
	// ReturnDocument(After)  вернуть документ уже после изменения
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated Contact
	err := contacts.FindOneAndUpdate(r.Context(), bson.D{{Key: "_id", Value: id}}, update, opts).Decode(&updated)
	if errors.Is(err, mongo.ErrNoDocuments) {
		handleError(w, "Нет контакта", "Контакт не найден", http.StatusNotFound)
		return
	}
	if err != nil {
		handleError(w, err.Error(), "Не удалось обновить контакт", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// DELETE /api/contacts/{id}  удалить контакт, вернуть его id
func deleteContact(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	res, err := contacts.DeleteOne(r.Context(), bson.D{{Key: "_id", Value: id}})
	if err != nil {
		handleError(w, err.Error(), "Не удалось удалить контакт", http.StatusInternalServerError)
		return
	}
	if res.DeletedCount == 0 {
		handleError(w, "Нет контакта", "Контакт не найден", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"_id": id.Hex()})
}

// spaHandler раздаёт собранный Angular; на неизвестные адреса отдаёт index.html
func spaHandler(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(dir, filepath.Clean("/"+r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			files.ServeHTTP(w, r) // существующий файл: js, css, favicon
			return
		}
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}

func main() {
	uri := getenv("MONGODB_URI", "mongodb://localhost:27017/phonebook")
	port := getenv("PORT", "8080")
	// Каталог сборки Angular из проекта pr5
	staticDir := getenv("STATIC_DIR", filepath.Join("..", "pr5", "client", "dist", "client", "browser"))

	// сначала подключаемся к базе, потом запускаем сервер
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Нет подключения к MongoDB: ", err)
	}
	dbName := "phonebook"
	if cs := strings.Split(strings.Split(uri, "?")[0], "/"); len(cs) > 3 && cs[3] != "" {
		dbName = cs[3] // имя базы из строки подключения
	}
	contacts = client.Database(dbName).Collection("contacts")
	log.Println("База данных подключена:", dbName)

	// маршруты метод + путь, {id}  параметр пути
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/contacts", listContacts)
	mux.HandleFunc("POST /api/contacts", createContact)
	mux.HandleFunc("DELETE /api/contacts", deleteAll)
	mux.HandleFunc("GET /api/contacts/{id}", getContact)
	mux.HandleFunc("PUT /api/contacts/{id}", updateContact)
	mux.HandleFunc("DELETE /api/contacts/{id}", deleteContact)
	mux.Handle("GET /", spaHandler(staticDir))

	log.Println("Приложение запущено на порту", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
