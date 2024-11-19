package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Book struct {
	ID         int
	Title      string
	Genres     []string
	AvgRating  float64
	NumRatings int
}

type Peti struct {
	Send     int
	Opc      int
	MovGenre []string
}

var allGenres []string

func loaDataset(ruta_1 string, ruta_2 string) ([]Book, error) {
	file_1, err := os.Open(ruta_1)
	if err != nil {
		return nil, err
	}
	file_2, err := os.Open(ruta_2)
	if err != nil {
		return nil, err
	}
	defer file_1.Close()
	defer file_2.Close()

	var books []Book
	reader_1 := csv.NewReader(file_1)
	reader_2 := csv.NewReader(file_2)
	_, _ = reader_1.Read()
	_, _ = reader_2.Read()

	for {
		record_1, err := reader_1.Read()
		if err != nil {
			break
		}
		id, err := strconv.Atoi(record_1[0])
		if err != nil {
			fmt.Println("Error converting ID:", err)
			continue
		}
		//genres := strings.Split(strings.Trim(record_1[2], "[]"), "|")
		genres := strings.Split(strings.ToLower(strings.Trim(record_1[2], "[]")), "|")
		books = append(books, Book{
			ID:     id,
			Title:  record_1[1],
			Genres: genres,
		})
		addGenres(genres)
	}

	dicc_rat := make(map[int]float64)
	dicc_num_rat := make(map[int]int)

	for {
		record_2, err := reader_2.Read()
		if err != nil {
			break
		}
		id, err := strconv.Atoi(record_2[1])
		if err != nil {
			fmt.Println("Error converting rating ID:", err)
			continue
		}
		rat, err := strconv.ParseFloat(record_2[2], 64)
		if err != nil {
			fmt.Println("Error converting rating:", err)
			continue
		}
		dicc_rat[id] += rat
		dicc_num_rat[id]++
	}

	for i := range books {
		id := books[i].ID
		if count, exists := dicc_num_rat[id]; exists && count > 0 {
			books[i].AvgRating = dicc_rat[id] / float64(count)
			books[i].NumRatings = count
		}
	}

	return books, nil
}

func addGenres(genres []string) {
	for _, genre := range genres {
		if !contains(allGenres, genre) {
			allGenres = append(allGenres, genre)
		}
	}
}

func genresToVector(genres []string) []int {
	vector := make([]int, len(allGenres))
	for i, genre := range allGenres {
		if contains(genres, genre) {
			vector[i] = 1
		} else {
			vector[i] = 0
		}
	}
	return vector
}

func cosineSimilarity(vec1, vec2 []int) float64 {
	var dotProduct, normA, normB float64
	for i := 0; i < len(vec1); i++ {
		dotProduct += float64(vec1[i] * vec2[i])
		normA += float64(vec1[i] * vec1[i])
		normB += float64(vec2[i] * vec2[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func recommendWithMultipleFactors(books []Book, targetGenres []string, excludedTitle string) []Book {
	targetVector := genresToVector(targetGenres)
	var recommendations []struct {
		Book  Book
		Score float64
	}

	for _, book := range books {
		// Excluir el libro con el título especificado
		if strings.EqualFold(book.Title, excludedTitle) {
			continue
		}

		bookVector := genresToVector(book.Genres)
		genreSimilarity := cosineSimilarity(targetVector, bookVector)

		commonGenresCount := countCommonGenres(book.Genres, targetGenres)
		percentageMatch := float64(commonGenresCount) / float64(len(targetGenres))

		if commonGenresCount == 0 {
			continue
		}

		const baseGenreWeight, ratingWeight, numRatingsWeight = 0.6, 0.25, 0.15
		//genreWeight := baseGenreWeight + (0.2 * float64(commonGenresCount))
		genreWeight := baseGenreWeight * percentageMatch

		score := (genreWeight * genreSimilarity) +
			(ratingWeight * (book.AvgRating / 5.0)) +
			(numRatingsWeight * (math.Log10(float64(book.NumRatings+1)) / 3.0))

		if score > 0 {
			recommendations = append(recommendations, struct {
				Book  Book
				Score float64
			}{Book: book, Score: score})
		}
	}

	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	var topRecommendations []Book
	for i, rec := range recommendations {
		if i >= 5 {
			break
		}
		topRecommendations = append(topRecommendations, rec.Book)
	}

	return topRecommendations
}

func countCommonGenres(bookGenres, targetGenres []string) int {
	count := 0
	for _, genre := range bookGenres {
		for _, target := range targetGenres {
			if strings.EqualFold(strings.TrimSpace(genre), strings.TrimSpace(target)) {
				count++
				break // Se evita duplicados en conteo para el mismo género
			}
		}
	}
	return count
}

func getGenresByTitle(books []Book, title string) []string {
	for _, book := range books {
		if strings.EqualFold(book.Title, title) {
			return book.Genres
		}
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func handle(con net.Conn, books []Book) {
	defer con.Close()

	/*msg, _ := bufio.NewReader(con).ReadString('\n')
	var petic Peti
	fmt.Println(msg)
	err := json.Unmarshal([]byte(strings.TrimSpace(msg)), &petic)
	if err != nil {
		fmt.Println("Error al deserializar JSON:", err)
		return
	}

	switch petic.Send {
	case 1: // Buscar géneros por título
		---var petic Peti
		err := json.Unmarshal([]byte(strings.TrimSpace(msg[1:])), &petic)
		if err != nil {
			fmt.Println("Error al deserializar JSON:", err)
			return
		}---
		fmt.Println("Recomendaciones por título")
		// Obtener géneros a partir del título de la película
		title := petic.MovGenre[0]
		fmt.Println("Título de la Película: ", title)
		genres := getGenresByTitle(books, title)
		if genres == nil {
			// Enviar mensaje de error si no se encuentra el libro
			errorMessage := "No se pudo encontrar el libro solicitado. Intente con la opción de géneros."
			fmt.Fprint(con, errorMessage+"\n")
			return
		}

		recommendations := recommendWithMultipleFactors(books, genres, title)

		// Enviar recomendaciones al maestro
		response, err := json.Marshal(recommendations)
		if err != nil {
			fmt.Println("Error al codificar JSON:", err)
			return
		}
		fmt.Fprint(con, string(response)+"\n")

	case 2: // Recomendaciones basadas en géneros proporcionados
		---var petic Peti
		err := json.Unmarshal([]byte(strings.TrimSpace(msg[1:])), &petic)
		if err != nil {
			fmt.Println("Error al deserializar JSON:", err)
			return
		}---
		fmt.Println("Recomendaciones por género")
		fmt.Println("Géneros recibidos para recomendaciones:", petic.MovGenre)

		recommendations := recommendWithMultipleFactors(books, petic.MovGenre, "")

		// Enviar recomendaciones al maestro
		response, err := json.Marshal(recommendations)
		if err != nil {
			fmt.Println("Error al codificar JSON:", err)
			return
		}
		fmt.Fprint(con, string(response)+"\n")
	}*/
	reader := bufio.NewReader(con)

	for {
		// Lee el mensaje del maestro
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error al leer mensaje:", err)
			break // Sale del ciclo y cierra la conexión
		}

		var petic Peti
		fmt.Println("Mensaje recibido:", msg)

		// Intenta deserializar el mensaje JSON
		err = json.Unmarshal([]byte(strings.TrimSpace(msg)), &petic)
		if err != nil {
			fmt.Println("Error al deserializar JSON:", err)
			continue // Sigue esperando nuevos mensajes
		}

		// Procesa la solicitud según el valor de `Send`
		switch petic.Send {
		case 1: // Recomendaciones por título
			fmt.Println("Recomendaciones por título")
			title := petic.MovGenre[0]
			fmt.Println("Título de la Película:", title)

			genres := getGenresByTitle(books, title)
			if genres == nil {
				errorMessage := "No se pudo encontrar el libro solicitado. Intente con la opción de géneros."
				fmt.Fprint(con, errorMessage+"\n")
				continue
			}

			recommendations := recommendWithMultipleFactors(books, genres, title)
			response, err := json.Marshal(recommendations)
			if err != nil {
				fmt.Println("Error al codificar JSON:", err)
				continue
			}
			fmt.Fprint(con, string(response)+"\n")

		case 2: // Recomendaciones basadas en géneros
			fmt.Println("Recomendaciones por género")
			fmt.Println("Géneros recibidos:", petic.MovGenre)

			recommendations := recommendWithMultipleFactors(books, petic.MovGenre, "")
			response, err := json.Marshal(recommendations)
			if err != nil {
				fmt.Println("Error al codificar JSON:", err)
				continue
			}
			fmt.Fprint(con, string(response)+"\n")
		}
	}
}

func active(books []Book) {
	ls, err := net.Listen("tcp", "0.0.0.0:9002")
	if err != nil {
		fmt.Println("Error al iniciar el trabajador:", err)
		return
	}
	defer ls.Close()

	fmt.Println("Trabajador escuchando en el puerto 9002...")

	for {
		con, err := ls.Accept()
		if err != nil {
			fmt.Println("Error al aceptar conexión:", err)
			continue
		}
		go handle(con, books)
	}
}

func main() {
	fmt.Println("Cargando dataset ...")
	books, err := loaDataset("/app/data/movies.csv", "/app/data/ratings.csv")
	fmt.Println("Dataset cargado :D")
	if err != nil {
		fmt.Println("Error cargando el dataset:", err)
		return
	}

	active(books)
}
