package controllers

import (
	"net/http"

	models "padi-template/app/Models"
	"github.com/wibiesana/padi_go_core/activerecord"
	"github.com/wibiesana/padi_go_core/query"
	"github.com/wibiesana/padi_go_core/response"
	"github.com/wibiesana/padi_go_core/router"
	"github.com/wibiesana/padi_go_core/validator"
)

type UserController struct{}

func NewUserController() *UserController {
	return &UserController{}
}

// Index lists records with pagination and search
func (c *UserController) Index(w http.ResponseWriter, r *http.Request) {
	opts := query.ParseOptions(r)
	searchColumns := []string{"name", "username", "email", "role"}
	meta, records, err := activerecord.PaginateWithContext[models.User](r.Context(), opts, searchColumns...)
	if err != nil {
		response.InternalServerError(w, "Failed to retrieve User list")
		return
	}

	response.Paginated(w, records, meta, "User list retrieved successfully")
}

// All retrieves all users without pagination (with optional search and sorting)
func (c *UserController) All(w http.ResponseWriter, r *http.Request) {
	opts := query.ParseOptions(r)
	searchColumns := []string{"name", "username", "email", "role"}
	records, err := activerecord.AllWithOptsWithContext[models.User](r.Context(), opts, searchColumns...)
	if err != nil {
		response.InternalServerError(w, "Failed to retrieve all users")
		return
	}

	response.Items(w, records, "All users retrieved successfully")
}

// Show retrieves a single record by ID
func (c *UserController) Show(w http.ResponseWriter, r *http.Request) {
	id, err := router.ParamUint(r, "id")
	if err != nil {
		response.BadRequest(w, "Invalid ID parameter")
		return
	}

	item, err := activerecord.FindOrFailWithContext[models.User](r.Context(), id)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	response.Item(w, item, "User retrieved successfully")
}

// Store creates a new record
func (c *UserController) Store(w http.ResponseWriter, r *http.Request) {
	var item models.User
	if errs, err := validator.BindJSON(r, &item); err != nil {
		response.UnprocessableEntity(w, errs, "Validation failed")
		return
	}

	if item.Password != "" {
		if err := item.SetPassword(item.Password); err != nil {
			response.InternalServerError(w, "Failed to hash password")
			return
		}
	}

	if err := activerecord.Save(&item, r.Context()); err != nil {
		response.InternalServerError(w, "Failed to create User: "+err.Error())
		return
	}

	response.Created(w, item, "User created successfully")
}

// Update updates an existing record
func (c *UserController) Update(w http.ResponseWriter, r *http.Request) {
	id, err := router.ParamUint(r, "id")
	if err != nil {
		response.BadRequest(w, "Invalid ID parameter")
		return
	}

	item, err := activerecord.FindOrFailWithContext[models.User](r.Context(), id)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	var updateData struct {
		Name     string `json:"name"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if errs, err := validator.BindJSON(r, &updateData); err != nil {
		response.UnprocessableEntity(w, errs, "Validation failed")
		return
	}

	if updateData.Name != "" {
		name := updateData.Name
		item.Name = &name
	}
	if updateData.Username != "" {
		username := updateData.Username
		item.Username = &username
	}
	if updateData.Email != "" {
		item.Email = updateData.Email
	}
	if updateData.Role != "" {
		item.Role = updateData.Role
	}
	if updateData.Password != "" {
		if err := item.SetPassword(updateData.Password); err != nil {
			response.InternalServerError(w, "Failed to hash password")
			return
		}
	}

	if err := activerecord.Save(item, r.Context()); err != nil {
		response.InternalServerError(w, "Failed to update User")
		return
	}

	response.Item(w, item, "User updated successfully")
}

// Destroy deletes a record
func (c *UserController) Destroy(w http.ResponseWriter, r *http.Request) {
	id, err := router.ParamUint(r, "id")
	if err != nil {
		response.BadRequest(w, "Invalid ID parameter")
		return
	}

	item, err := activerecord.FindOrFailWithContext[models.User](r.Context(), id)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	if err := activerecord.DeleteModel(item, r.Context()); err != nil {
		response.InternalServerError(w, "Failed to delete User")
		return
	}

	response.Success(w, nil, "User deleted successfully")
}
