package util

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
)

func ParseDatabaseError(err error, prefix string) error {
	var pgErr *pgconn.PgError
	errors.As(err, &pgErr)
	if pgErr.Code == "23505" {
		field := strings.Replace(pgErr.ConstraintName, prefix, "", -1)
		field = strings.Replace(field, "_", " ", -1)
		return fmt.Errorf("%s already exists", field)
	} else if pgErr.Code == "23503" {
		return fmt.Errorf("Related data not found")
	}
	return err
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
