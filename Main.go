package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

const filePath = "data.json"

type Car struct {
	ID          int    `json:"id"`
	Brand       string `json:"brand"`
	Model       string `json:"model"`
	Mileage     int    `json:"mileage"`
	OwnersCount int    `json:"owners_count"`
}

type Data struct {
	Cars []Car `json:"cars"`
}

var data Data

// Чтение данных из файла
func loadData() error {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Файл не существует, ничего страшного
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&data)
}

// Запись данных в файл
func saveData() error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	message := `
	REST API is working!
	Available routes:
	- GET /cars            -> get a list of cars`
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(message))
}

func addCar(w http.ResponseWriter, r *http.Request) {
	var car Car
	if err := json.NewDecoder(r.Body).Decode(&car); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	car.ID = len(data.Cars) + 1
	data.Cars = append(data.Cars, car)
	saveData()
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(car)
}

func deleteCar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	for i, car := range data.Cars {
		if car.ID == id {
			data.Cars = append(data.Cars[:i], data.Cars[i+1:]...)
			saveData()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Car deleted successfully"))
			return
		}
	}
	http.Error(w, "Car not found", http.StatusNotFound)
}

func updateCar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updatedCar Car
	if err := json.NewDecoder(r.Body).Decode(&updatedCar); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for i, car := range data.Cars {
		if car.ID == id {
			updatedCar.ID = id
			data.Cars[i] = updatedCar
			saveData()
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(updatedCar)
			return
		}
	}
	http.Error(w, "Car not found", http.StatusNotFound)
}

func getCarByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	for _, car := range data.Cars {
		if car.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(car)
			return
		}
	}

	http.Error(w, "Car not found", http.StatusNotFound)
}

func patchCar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for i, car := range data.Cars {
		if car.ID == id {
			if brand, ok := updates["brand"].(string); ok {
				car.Brand = brand
			}
			if model, ok := updates["model"].(string); ok {
				car.Model = model
			}
			if mileage, ok := updates["mileage"].(float64); ok { // JSON числа по умолчанию float64
				car.Mileage = int(mileage)
			}
			if ownersCount, ok := updates["owners_count"].(float64); ok {
				car.OwnersCount = int(ownersCount)
			}

			data.Cars[i] = car
			saveData()

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(car)
			return
		}
	}

	http.Error(w, "Car not found", http.StatusNotFound)
}

func main() {
	// Загружаем данные из файла
	if err := loadData(); err != nil {
		fmt.Println("Error loading data:", err)
		return
	}

	router := mux.NewRouter()

	// Маршруты API
	router.HandleFunc("/", homeHandler).Methods("GET")

	//Списки
	router.HandleFunc("/cars", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data.Cars)
	}).Methods("GET")

	router.HandleFunc("/add_car", addCar).Methods("POST")
	router.HandleFunc("/cars/{id}", deleteCar).Methods("DELETE")
	router.HandleFunc("/cars/{id}", updateCar).Methods("PUT")
	router.HandleFunc("/cars/{id}", getCarByID).Methods("GET")
	router.HandleFunc("/cars/{id}", patchCar).Methods("PATCH")

	fmt.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", router)
}
