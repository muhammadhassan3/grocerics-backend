package util

import (
	"errors"
	"fmt"
	"strings"

	"grocerics-backend/internal/errs"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
)

func ParseDatabaseError(err error, prefix string) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	if pgErr.Code == "23505" {
		field := strings.TrimSpace(strings.ReplaceAll(strings.Replace(pgErr.ConstraintName, prefix, "", 1), "_", " "))
		if field == "slug" {
			field = "name"
		}
		resource := resourceFromIndexPrefix(prefix)
		if resource == "" {
			return errs.Conflict("ALREADY_EXISTS", fmt.Sprintf("This %s already exists", field))
		}
		return errs.Conflict("ALREADY_EXISTS", fmt.Sprintf("A %s with this %s already exists", resource, field))
	} else if pgErr.Code == "23503" {
		return errs.BadRequest("RELATED_NOT_FOUND", "Related data not found")
	}
	return err
}

func resourceFromIndexPrefix(prefix string) string {
	name := strings.Trim(strings.TrimPrefix(prefix, "idx_"), "_")
	name = strings.ReplaceAll(name, "_", " ")
	switch {
	case name == "":
		return ""
	case strings.HasSuffix(name, "ies"):
		return strings.TrimSuffix(name, "ies") + "y"
	case strings.HasSuffix(name, "s"):
		return strings.TrimSuffix(name, "s")
	}
	return name
}

func ParseValidationError(err error) error {
	var validationError validator.ValidationErrors
	if errors.As(err, &validationError) {
		var errorMessages []string
		for _, fieldErr := range validationError {
			errorMessages = append(errorMessages, fmt.Sprintf("%s is %s", fieldErr.Field(), fieldErr.Tag()))
		}
		return errors.New(strings.Join(errorMessages, ", "))
	}
	return err
}
