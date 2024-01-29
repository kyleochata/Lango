package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Task represents a task in the system
type Task struct {
	ID        string    `json:"id,omitempty" bson:"_id,omitempty"`
	Name      string    `json:"name,omitempty" bson:"name,omitempty"`
	Completed bool      `json:"completed,omitempty" bson:"completed,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

var client *mongo.Client

func initDB() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	var err error
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
}

func getCollection() *mongo.Collection {
	return client.Database("mydatabase").Collection("tasks")
}

func GetTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var tasks []Task
	collection := getCollection()
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var task Task
		cursor.Decode(&task)
		tasks = append(tasks, task)
	}
	json.NewEncoder(w).Encode(tasks)
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	var task Task
	collection := getCollection()
	err := collection.FindOne(context.Background(), bson.M{"_id": params["id"]}).Decode(&task)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(task)
}

func CreateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var task Task
	_ = json.NewDecoder(r.Body).Decode(&task)
	task.ID = primitive.NewObjectID().Hex()
	task.CreatedAt = time.Now()
	collection := getCollection()
	_, err := collection.InsertOne(context.Background(), task)
	if err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(task)
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	var updatedTask Task
	_ = json.NewDecoder(r.Body).Decode(&updatedTask)
	collection := getCollection()
	filter := bson.M{"_id": params["id"]}
	update := bson.M{"$set": updatedTask}
	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	updatedTask.ID = params["id"]
	json.NewEncoder(w).Encode(updatedTask)
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	collection := getCollection()
	_, err := collection.DeleteOne(context.Background(), bson.M{"_id": params["id"]})
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(bson.M{"message": "Task deleted successfully"})
}

func InitializeRoutes(router *mux.Router) {
	router.HandleFunc("/tasks", GetTasks).Methods("GET")
	router.HandleFunc("/tasks/{id}", GetTask).Methods("GET")
	router.HandleFunc("/tasks", CreateTask).Methods("POST")
	router.HandleFunc("/tasks/{id}", UpdateTask).Methods("PUT")
	router.HandleFunc("/tasks/{id}", DeleteTask).Methods("DELETE")
}

func main() {
	router := mux.NewRouter()
	InitializeRoutes(router)
	log.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
