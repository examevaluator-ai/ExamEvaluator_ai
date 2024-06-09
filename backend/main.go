package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var (
	processingStatus string
	mu               sync.Mutex
	transcript       map[string]interface{}
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/upload", UploadHandler).Methods("POST")
	r.HandleFunc("/status", StatusHandler).Methods("GET")
	r.HandleFunc("/result", ResultHandler).Methods("GET")

	// Configurar CORS
	corsObj := handlers.AllowedOrigins([]string{"http://localhost:3000"})
	methods := handlers.AllowedMethods([]string{"POST", "GET"})
	headers := handlers.AllowedHeaders([]string{"Content-Type"})

	fmt.Println("Server started at :8080")
	if err := http.ListenAndServe(":8080", handlers.CORS(corsObj, methods, headers)(r)); err != nil {
		fmt.Println("Failed to start server", err)
	}
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	status := processingStatus
	mu.Unlock()
	w.Write([]byte(status))
}

func ResultHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	result, err := json.Marshal(transcript)
	mu.Unlock()
	if err != nil {
		http.Error(w, "Error fetching the transcript", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	file, handler, err := r.FormFile("video")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Caminho absoluto para a pasta de uploads
	uploadDir := "C:/Users/diego/Desktop/ExamEvaluator.ai/backend/uploads"

	// Cria a pasta de uploads se não existir
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, os.ModePerm)
	}

	// Define o caminho para salvar o arquivo
	filePath := filepath.Join(uploadDir, handler.Filename)

	// Cria o arquivo no disco
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// Escreve o conteúdo do arquivo recebido no arquivo criado no disco
	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Atualiza o status de upload completo
	mu.Lock()
	processingStatus = "transcribing audio"
	mu.Unlock()

	// Extrair áudio do vídeo
	audioPath := filepath.Join(uploadDir, "audio.wav")
	err = extractAudio(filePath, audioPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Chamar script Python para processamento
	go runPythonScript(audioPath)

	fmt.Fprintln(w, "Upload successful, processing started.")
}

func extractAudio(videoPath, audioPath string) error {
	cmd := exec.Command("ffmpeg", "-y", "-i", videoPath, "-q:a", "0", "-map", "a", audioPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("Running command: %s\n", cmd.String())
	return cmd.Run()
}

func runPythonScript(audioPath string) {
	venvPython := filepath.Join("C:/Users/diego/Desktop/ExamEvaluator.ai/ml/venv/Scripts/python.exe")
	cmd := exec.Command(venvPython, "C:/Users/diego/Desktop/ExamEvaluator.ai/ml/main.py", audioPath)

	mu.Lock()
	processingStatus = "transcribing audio"
	mu.Unlock()

	stdout, err := cmd.Output()
	if err != nil {
		mu.Lock()
		processingStatus = "error"
		mu.Unlock()
		fmt.Printf("Processing error: %v\n", err)
		fmt.Printf("Stdout: %s\n", string(stdout))
		return
	}

	// Lendo a transcrição do arquivo JSON
	jsonPath := filepath.Join(filepath.Dir(audioPath), filepath.Base(audioPath)[:len(filepath.Base(audioPath))-len(filepath.Ext(audioPath))]+".json")
	transcriptFile, err := os.Open(jsonPath)
	if err != nil {
		mu.Lock()
		processingStatus = "error"
		mu.Unlock()
		fmt.Printf("Error opening JSON transcript file: %v\n", err)
		return
	}
	defer transcriptFile.Close()

	decoder := json.NewDecoder(transcriptFile)
	err = decoder.Decode(&transcript)
	if err != nil {
		mu.Lock()
		processingStatus = "error"
		mu.Unlock()
		fmt.Printf("Error decoding JSON transcript: %v\n", err)
		return
	}

	mu.Lock()
	processingStatus = "done"
	fmt.Printf("Transcript result: %v\n", transcript)
	mu.Unlock()
}
