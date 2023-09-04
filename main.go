package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/gin-gonic/gin"
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

type Missa struct {
	Name     string `json:"name"`
	Date     string `json:"date"`
	Donations  []Donation `json:"donations"`
}

var missas []Missa
var novenas []Novena
var currentNovenaID int
var currentMissaID int

func main() {
	loadNovenasFromFile()

	router := gin.Default()

	/*desativado por enquanto*/
	//router.GET("/", donationStore)
	//router.POST("/", donationStore)

	router.GET("/", home)

	router.GET("/list", listDonations)
	router.GET("/novena-cad", novenaStore)
	router.POST("/novena-cad", novenaStore)
	router.GET("/print-novena/:id", detailNovena)

	router.Run(":3000")
	//router.Run()

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

func serveHTML(c *gin.Context, content string) {
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
						<!--<li class="nav-item">
							<a class="nav-link" href="/novena-cad">Cadastrar Novena</a>
						</li>
						<li class="nav-item">
							<a class="nav-link" href="/list">Ver Doações</a>
						</li>-->
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

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func home(c *gin.Context){
	content := `
		<h1 class="mb-4">Bem vido!</h1>
	`

	serveHTML(c, content)
}

func donationStore(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		novenaName := c.PostForm("novena")
		name := c.PostForm("name")
		donation := c.PostForm("donation")
		date := c.PostForm("date")

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

	serveHTML(c, content)
}

func listDonations(c *gin.Context) {
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
	c.Header("Content-Type", "text/html; charset=utf-8")
	tmplParsed.Execute(c.Writer, novenas)
}

func detailNovena(c *gin.Context) {
	id := c.Param("id")
	novenaID, err := strconv.Atoi(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid Novena ID")
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
		c.String(http.StatusNotFound, "Novena not found")
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
	c.Header("Content-Type", "text/html; charset=utf-8")
	tmplParsed.Execute(c.Writer, targetNovena)
}

func novenaStore(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		novenaName := c.PostForm("novena")
		dateinicio := c.PostForm("dateinicio")
		datefim := c.PostForm("datefim")

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

	serveHTML(c, content)
}
