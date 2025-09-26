package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

const (
	ModelsDir   = "internal/models"
	ReposDir    = "internal/repositories"
	ServicesDir = "internal/services"
	HandlersDir = "internal/handlers"
)

// Field representa un campo del modelo
type Field struct {
	Name    string
	Type    string
	GormTag string
	JsonTag string
}

// ModelData contiene la información del modelo
type ModelData struct {
	Name          string
	LowerName     string
	Fields        []Field
	HasTimestamps bool
}

// Templates
const modelTemplate = `package models

import (
	"time"
)

type {{.Name}} struct {
{{if .HasTimestamps}}	ID        uint      ` + "`gorm:\"primaryKey\" json:\"id\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `

{{end}}{{range .Fields}}	{{.Name}} {{.Type}} ` + "`{{.JsonTag}}{{if .GormTag}} {{.GormTag}}{{end}}`" + `
{{end}}}
`

const repositoryTemplate = `package repositories

import (
	"log"
	"time"

	"github.com/imlargo/go-api/internal/models"
	"gorm.io/gorm/clause"
)

type {{.Name}}Repository interface {
	Create({{.LowerName}} *models.{{.Name}}) error
	GetAll() ([]*models.{{.Name}}, error)
	Get(id uint) (*models.{{.Name}}, error)
	Update({{.LowerName}} *models.{{.Name}}) error
	Patch(id uint, data *map[string]interface{}) error
	Delete(id uint) error
}

type {{.LowerName}}Repository struct {
	*Repository
}

func New{{.Name}}Repository(
	r *Repository,
) {{.Name}}Repository {
	return &{{.LowerName}}Repository{
		Repository: r,
	}
}

func (r *{{.LowerName}}Repository) Create({{.LowerName}} *models.{{.Name}}) error {
	return r.db.Create({{.LowerName}}).Error
}

func (r *{{.LowerName}}Repository) Get(id uint) (*models.{{.Name}}, error) {
	var {{.LowerName}} models.{{.Name}}

	if err := r.db.First(&{{.LowerName}}, id).Error; err != nil {
		return nil, err
	}

	return &{{.LowerName}}, nil
}

func (r *{{.LowerName}}Repository) Update({{.LowerName}} *models.{{.Name}}) error {
	if err := r.db.Model({{.LowerName}}).Clauses(clause.Returning{}).Updates({{.LowerName}}).Error; err != nil {
		return err
	}

	return nil
}

func (r *{{.LowerName}}Repository) Patch(id uint, data *map[string]interface{}) error {
	if err := r.db.Model(&models.{{.Name}}{}).Where("id = ?", id).Updates(data).Error; err != nil {
		return err
	}

	return nil
}

func (r *{{.LowerName}}Repository) Delete(id uint) error {
	if err := r.db.Delete(&models.{{.Name}}{}, id).Error; err != nil {
		return err
	}

	return nil
}

func (r *{{.LowerName}}Repository) GetAll() ([]*models.{{.Name}}, error) {
	var {{.LowerName}}s []*models.{{.Name}}
	if err := r.db.Find(&{{.LowerName}}s).Error; err != nil {
		return nil, err
	}
	return {{.LowerName}}s, nil
}
`

const serviceTemplate = `package services

import (
	"github.com/imlargo/go-api/internal/models"
)

type {{.Name}}Service interface {
	Create{{.Name}}({{.LowerName}} *models.{{.Name}}) (*models.{{.Name}}, error)
	Delete{{.Name}}({{.LowerName}}ID uint) error
	Update{{.Name}}({{.LowerName}}ID uint, data *models.{{.Name}}) (*models.{{.Name}}, error)
	Get{{.Name}}ByID({{.LowerName}}ID uint) (*models.{{.Name}}, error)
	GetAll{{.Name}}s() ([]*models.{{.Name}}, error)
}

type {{.LowerName}}Service struct {
	*Service
}

func New{{.Name}}Service(service *Service) {{.Name}}Service {
	return &{{.LowerName}}Service{
		Service: service,
	}
}

func (s *{{.LowerName}}Service) Create{{.Name}}({{.LowerName}} *models.{{.Name}}) (*models.{{.Name}}, error) {
	if err := s.store.{{.Name}}s.Create({{.LowerName}}); err != nil {
		return nil, err
	}

	return {{.LowerName}}, nil
}

func (s *{{.LowerName}}Service) Delete{{.Name}}({{.LowerName}}ID uint) error {
	return s.store.{{.Name}}s.Delete({{.LowerName}}ID)
}

func (s *{{.LowerName}}Service) Update{{.Name}}({{.LowerName}}ID uint, data *models.{{.Name}}) (*models.{{.Name}}, error) {
	existing{{.Name}}, err := s.store.{{.Name}}s.GetByID({{.LowerName}}ID)
	if err != nil {
		return nil, err
	}

	// Update fields
	data.ID = existing{{.Name}}.ID
	data.CreatedAt = existing{{.Name}}.CreatedAt

	if err := s.store.{{.Name}}s.Update(data); err != nil {
		return nil, err
	}

	return data, nil
}

func (s *{{.LowerName}}Service) Get{{.Name}}ByID({{.LowerName}}ID uint) (*models.{{.Name}}, error) {
	return s.store.{{.Name}}s.GetByID({{.LowerName}}ID)
}

func (s *{{.LowerName}}Service) GetAll{{.Name}}s() ([]*models.{{.Name}}, error) {
	return s.store.{{.Name}}s.GetAll()
}
`

const handlerTemplate = `package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/internal/models"
	"github.com/imlargo/go-api/internal/responses"
)

type {{.Name}}Handler struct {
	*Handler
}

func New{{.Name}}Handler(handler *Handler) *{{.Name}}Handler {
	return &{{.Name}}Handler{
		Handler: handler,
	}
}

// Create{{.Name}} creates a new {{.LowerName}}
// @Summary Create {{.LowerName}}
// @Description Create a new {{.LowerName}}
// @Tags {{.Name}}
// @Accept json
// @Produce json
// @Param {{.LowerName}} body models.{{.Name}} true "{{.Name}} data"
// @Success 201 {object} models.{{.Name}}
// @Router /{{.LowerName}}s [post]
func (h *{{.Name}}Handler) Create{{.Name}}(c *gin.Context) {
	var {{.LowerName}} models.{{.Name}}
	if err := c.ShouldBindJSON(&{{.LowerName}}); err != nil {
		responses.BadRequest(c, err.Error())
		return
	}

	created{{.Name}}, err := h.services.{{.Name}}s.Create{{.Name}}(&{{.LowerName}})
	if err != nil {
		responses.InternalError(c, err.Error())
		return
	}

	responses.Created(c, created{{.Name}})
}

// Get{{.Name}}ByID retrieves a {{.LowerName}} by ID
// @Summary Get {{.LowerName}} by ID
// @Description Get a specific {{.LowerName}} by its ID
// @Tags {{.Name}}
// @Produce json
// @Param id path int true "{{.Name}} ID"
// @Success 200 {object} models.{{.Name}}
// @Router /{{.LowerName}}s/{id} [get]
func (h *{{.Name}}Handler) Get{{.Name}}ByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		responses.BadRequest(c, "Invalid ID format")
		return
	}

	{{.LowerName}}, err := h.services.{{.Name}}s.Get{{.Name}}ByID(uint(id))
	if err != nil {
		responses.NotFound(c, "{{.Name}} not found")
		return
	}

	responses.Ok(c, {{.LowerName}})
}

// GetAll{{.Name}}s retrieves all {{.LowerName}}s
// @Summary Get all {{.LowerName}}s
// @Description Get all {{.LowerName}}s
// @Tags {{.Name}}
// @Produce json
// @Success 200 {array} models.{{.Name}}
// @Router /{{.LowerName}}s [get]
func (h *{{.Name}}Handler) GetAll{{.Name}}s(c *gin.Context) {
	{{.LowerName}}s, err := h.services.{{.Name}}s.GetAll{{.Name}}s()
	if err != nil {
		responses.InternalError(c, err.Error())
		return
	}

	responses.Ok(c, {{.LowerName}}s)
}

// Update{{.Name}} updates a {{.LowerName}}
// @Summary Update {{.LowerName}}
// @Description Update an existing {{.LowerName}}
// @Tags {{.Name}}
// @Accept json
// @Produce json
// @Param id path int true "{{.Name}} ID"
// @Param {{.LowerName}} body models.{{.Name}} true "{{.Name}} data"
// @Success 200 {object} models.{{.Name}}
// @Router /{{.LowerName}}s/{id} [put]
func (h *{{.Name}}Handler) Update{{.Name}}(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		responses.BadRequest(c, "Invalid ID format")
		return
	}

	var {{.LowerName}} models.{{.Name}}
	if err := c.ShouldBindJSON(&{{.LowerName}}); err != nil {
		responses.BadRequest(c, err.Error())
		return
	}

	updated{{.Name}}, err := h.services.{{.Name}}s.Update{{.Name}}(uint(id), &{{.LowerName}})
	if err != nil {
		responses.InternalError(c, err.Error())
		return
	}

	responses.Ok(c, updated{{.Name}})
}

// Delete{{.Name}} deletes a {{.LowerName}}
// @Summary Delete {{.LowerName}}
// @Description Delete a {{.LowerName}} by ID
// @Tags {{.Name}}
// @Param id path int true "{{.Name}} ID"
// @Success 204
// @Router /{{.LowerName}}s/{id} [delete]
func (h *{{.Name}}Handler) Delete{{.Name}}(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		responses.BadRequest(c, "Invalid ID format")
		return
	}

	if err := h.services.{{.Name}}s.Delete{{.Name}}(uint(id)); err != nil {
		responses.InternalError(c, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
`

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <command> <name>")
		fmt.Println("Commands:")
		fmt.Println("  model <name>   - Create model and repository")
		fmt.Println("  service <name> - Create service")
		fmt.Println("  handler <name> - Create handler")
		os.Exit(1)
	}

	command := os.Args[1]
	name := capitalizeFirst(os.Args[2])

	switch command {
	case "model":
		if err := createModel(name); err != nil {
			fmt.Printf("Error creating model: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Model %s and repository created successfully!\n", name)
	case "service":
		if err := createService(name); err != nil {
			fmt.Printf("Error creating service: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Service %s created successfully!\n", name)
	case "handler":
		if err := createHandler(name); err != nil {
			fmt.Printf("Error creating handler: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Handler %s created successfully!\n", name)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func createModel(name string) error {
	// Crear directorios si no existen
	if err := os.MkdirAll(ModelsDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(ReposDir, 0755); err != nil {
		return err
	}

	// Obtener campos del usuario
	fields, hasTimestamps := getFieldsFromUser()

	modelData := ModelData{
		Name:          name,
		LowerName:     strings.ToLower(name),
		Fields:        fields,
		HasTimestamps: hasTimestamps,
	}

	// Crear modelo
	modelPath := filepath.Join(ModelsDir, strings.ToLower(name)+".go")
	if err := createFileFromTemplate(modelPath, modelTemplate, modelData); err != nil {
		return err
	}

	// Crear repositorio
	repoPath := filepath.Join(ReposDir, strings.ToLower(name)+".go")
	if err := createFileFromTemplate(repoPath, repositoryTemplate, modelData); err != nil {
		return err
	}

	return nil
}

func createService(name string) error {
	if err := os.MkdirAll(ServicesDir, 0755); err != nil {
		return err
	}

	modelData := ModelData{
		Name:      name,
		LowerName: strings.ToLower(name),
	}

	servicePath := filepath.Join(ServicesDir, strings.ToLower(name)+".go")
	return createFileFromTemplate(servicePath, serviceTemplate, modelData)
}

func createHandler(name string) error {
	if err := os.MkdirAll(HandlersDir, 0755); err != nil {
		return err
	}

	modelData := ModelData{
		Name:      name,
		LowerName: strings.ToLower(name),
	}

	handlerPath := filepath.Join(HandlersDir, strings.ToLower(name)+".go")
	return createFileFromTemplate(handlerPath, handlerTemplate, modelData)
}

func getFieldsFromUser() ([]Field, bool) {
	scanner := bufio.NewScanner(os.Stdin)
	var fields []Field
	hasTimestamps := false

	fmt.Println("¿Incluir timestamps (ID, CreatedAt, UpdatedAt)? (y/n):")
	scanner.Scan()
	if strings.ToLower(scanner.Text()) == "y" || strings.ToLower(scanner.Text()) == "yes" {
		hasTimestamps = true
	}

	fmt.Println("Ingresa los campos del modelo (formato: nombre:tipo:gorm_tag:json_tag)")
	fmt.Println("Ejemplos:")
	fmt.Println("  name:string:not null:json:\"name\"")
	fmt.Println("  email:string:unique;not null:json:\"email\"")
	fmt.Println("  password:string:not null:json:\"-\"")
	fmt.Println("Presiona Enter sin texto para terminar:")

	for {
		fmt.Print("> ")
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			break
		}

		field, err := parseField(input)
		if err != nil {
			fmt.Printf("Error: %v. Intenta de nuevo.\n", err)
			continue
		}

		fields = append(fields, field)
	}

	return fields, hasTimestamps
}

func parseField(input string) (Field, error) {
	parts := strings.Split(input, ":")
	if len(parts) < 2 {
		return Field{}, fmt.Errorf("formato inválido. Usa: nombre:tipo[:gorm_tag[:json_tag]]")
	}

	field := Field{
		Name: capitalizeFirst(parts[0]),
		Type: parts[1],
	}

	// JSON tag por defecto
	field.JsonTag = fmt.Sprintf("json:\"%s\"", strings.ToLower(parts[0]))

	// Si se proporciona gorm tag
	if len(parts) > 2 && parts[2] != "" {
		field.GormTag = fmt.Sprintf("gorm:\"%s\"", parts[2])
	}

	// Si se proporciona json tag personalizado
	if len(parts) > 3 && parts[3] != "" {
		field.JsonTag = parts[3]
	}

	return field, nil
}

func createFileFromTemplate(path, tmpl string, data ModelData) error {
	t, err := template.New("template").Parse(tmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, data)
}

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}

	// Usar regex para encontrar la primera letra
	re := regexp.MustCompile(`^([a-z])`)
	return re.ReplaceAllStringFunc(s, strings.ToUpper)
}
