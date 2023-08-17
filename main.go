package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"text/template"
)

type Donation struct {
	Name     string `json:"name"`
	Donation string `json:"donation"`
	Date     string `json:"date"`
}

type Novena struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	DataInicio string     `json:"dateinicio"`
	DataFim    string     `json:"datefim"`
	Donations  []Donation `json:"donations"`
}

var novenas []Novena
var currentNovenaID int

func main() {
	loadNovenasFromFile()

	http.HandleFunc("/", donationStore)
	http.HandleFunc("/list", listDonations)
	http.HandleFunc("/novena-cad", novenaStore)
	http.HandleFunc("/print-novena/", detailNovena)
	http.ListenAndServe(":9080", nil)

	saveNovenasToFile()
}

func loadNovenasFromFile() {
	file, err := os.Open("novenas.json")
	if err != nil {
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&novenas)
	if err != nil {
		fmt.Println("Error loading novenas:", err)
	}
}

func saveNovenasToFile() {
	file, err := os.Create("novenas.json")
	if err != nil {
		fmt.Println("Error creating novenas file:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(novenas)
	if err != nil {
		fmt.Println("Error saving novenas:", err)
	}
}

func serveHTML(w http.ResponseWriter, r *http.Request, content string) {
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="utf-8">
			<title>Novena App</title>
			<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
		</head>
		<body>
			<nav class="navbar navbar-expand-lg navbar-light bg-light">
				<div class="container">
					<a class="navbar-brand" href="/">Novena App</a>
					<ul class="navbar-nav ml-auto">
						<li class="nav-item">
							<a class="nav-link" href="/novena-cad">Cadastrar Novena</a>
						</li>
						<li class="nav-item">
							<a class="nav-link" href="/list">Ver Doações</a>
						</li>
					</ul>
				</div>
			</nav>
			<div class="container mt-4">
				` + content + `
			</div>
			<script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
		</body>
		</html>
	`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func donationStore(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		novenaName := r.FormValue("novena")
		name := r.FormValue("name")
		donation := r.FormValue("donation")
		date := r.FormValue("date")

		var targetNovena *Novena
		for i := range novenas {
			if novenas[i].Name == novenaName {
				targetNovena = &novenas[i]
				break
			}
		}

		if targetNovena == nil {
			currentNovenaID++
			targetNovena = &Novena{
				ID:   currentNovenaID,
				Name: novenaName,
			}
			novenas = append(novenas, *targetNovena)
		}

		newDonation := Donation{
			Name:     name,
			Donation: donation,
			Date:     date,
		}

		targetNovena.Donations = append(targetNovena.Donations, newDonation)
		saveNovenasToFile()
	}

	content := `
		<h1 class="mb-4">Cadastro de Doação</h1>
		<form method="post" action="/">
			<label for="novena">Novena:</label>
			<input type="text" name="novena" required class="form-control mb-2">
			<label for="name">Nome:</label>
			<input type="text" name="name" required class="form-control mb-2">
			<label for="donation">Doação:</label>
			<input type="text" name="donation" required class="form-control mb-2">
			<label for="date">Data:</label>
			<input type="date" name="date" required class="form-control mb-2">
			<input type="submit" value="Cadastrar" class="btn btn-primary">
		</form>
	`

	serveHTML(w, r, content)
}

func listDonations(w http.ResponseWriter, r *http.Request) {
	tmpl := `
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="utf-8">
			<title>Lista de Doações</title>
			<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
			<style>
				.toggle-list {
					cursor: pointer;
					text-decoration: underline;
					color: blue;
				}
				.hidden {
					display: none;
				}
			</style>
			<script>
				function toggleList(id) {
					var list = document.getElementById("list-" + id);
					list.classList.toggle("hidden");
				}
			</script>
		</head>
		<body>
			<nav class="navbar navbar-expand-lg navbar-light bg-light">
				<div class="container">
					<a class="navbar-brand" href="/">Novena App</a>
					<ul class="navbar-nav ml-auto">
						<li class="nav-item">
							<a class="nav-link" href="/novena-cad">Cadastrar Novena</a>
						</li>
						<li class="nav-item">
							<a class="nav-link" href="/list">Ver Doações</a>
						</li>
					</ul>
				</div>
			</nav>
			<div class="container mt-4">
				<h1 class="mb-4">Lista de Doações</h1>
				{{range .}}
					<h2>
						<span class="toggle-list" onclick="toggleList({{.ID}})">Novena (ID {{.ID}}): <strong>{{.Name}}</strong> - Inicio: {{.DataInicio}} - Fim: {{.DataFim}}</span>
						<a href="/print-novena/{{.ID}}" class="btn btn-secondary btn-sm ml-2">visualizar</a>
					</h2>
					<ul id="list-{{.ID}}" class="hidden">
						{{range .Donations}}
						<li>
							<strong>Nome:</strong> {{.Name}}, 
							<strong>Doação:</strong> {{.Donation}}, 
							<strong>Data:</strong> {{.Date}}
						</li>
						{{end}}
					</ul>
				{{end}}
			</div>
			<script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
		</body>
		</html>
	`

	tmplParsed := template.Must(template.New("").Parse(tmpl))
	w.Header().Set("Content-Type", "text/html")
	tmplParsed.Execute(w, novenas)
}

func detailNovena(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/print-novena/"):]
	novenaID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid Novena ID", http.StatusBadRequest)
		return
	}

	var targetNovena *Novena
	for i := range novenas {
		if novenas[i].ID == novenaID {
			targetNovena = &novenas[i]
			break
		}
	}

	if targetNovena == nil {
		http.Error(w, "Novena not found", http.StatusNotFound)
		return
	}

	tmpl := `
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="utf-8">
			<title>Impressão da Novena - {{.Name}}</title>
			<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
		</head>
		<body>
			<div class="container mt-4">
				<h1 class="mb-4">Novena - {{.Name}}</h1>
				<h2>Início: {{.DataInicio}} - Fim: {{.DataFim}}</h2>
				<ul>
					{{range .Donations}}
					<li>
						<strong>Nome:</strong> {{.Name}}, 
						<strong>Doação:</strong> {{.Donation}}, 
						<strong>Data:</strong> {{.Date}}
					</li>
					{{end}}
				</ul>
			</div>
			<div class="container mt-4">
				<button class="btn btn-primary" id="printButton" onclick="printPage()">Imprimir</button>
				<a href="/list" class="btn btn-secondary" id="voltarButton">Voltar</a>
			</div>
			<script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
			<script>
				function printPage() {
					var printButton = document.getElementById("printButton");
					printButton.style.display = "none";

					var voltarButton = document.getElementById("voltarButton");
					voltarButton.style.display = "none";

					window.print();

					setTimeout(() => {
						printButton.style.display = "block";
						voltarButton.style.display = "block";
					}, 200);
				}
			</script>
		</body>
		</html>
	`

	tmplParsed := template.Must(template.New("").Parse(tmpl))
	w.Header().Set("Content-Type", "text/html")
	tmplParsed.Execute(w, targetNovena)
}

func novenaStore(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		novenaName := r.FormValue("novena")
		dateinicio := r.FormValue("dateinicio")
		datefim := r.FormValue("datefim")

		var targetNovena *Novena
		for i := range novenas {
			if novenas[i].Name == novenaName {
				targetNovena = &novenas[i]
				break
			}
		}

		if targetNovena == nil {
			currentNovenaID++
			targetNovena = &Novena{
				ID:         currentNovenaID,
				Name:       novenaName,
				DataInicio: dateinicio,
				DataFim:    datefim,
			}
			novenas = append(novenas, *targetNovena)
		}
	}

	content := `
		<h1 class="mb-4">Cadastro da Novena</h1>
		<form method="post" action="/novena-cad">
			<label for="novena">Novena:</label>
			<input type="text" name="novena" required class="form-control mb-2">
			<label for="date">Data Inicio:</label>
			<input type="date" name="dateinicio" required class="form-control mb-2">
			<label for="date">Data Fim:</label>
			<input type="date" name="datefim" required class="form-control mb-2">
			<input type="submit" value="Cadastrar" class="btn btn-primary">
		</form>
	`

	serveHTML(w, r, content)
}
