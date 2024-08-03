package utils

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
)

type EmailTemplateDetails struct {
	Subject      string
	RedirectURL  string
	BodyTemplate string
}

func GetVerifyEmailDetails(db *sql.DB) (EmailTemplateDetails, error) {
	emailTemplate := EmailTemplateDetails{}

	rows, err := db.Query("SELECT name, value FROM settings WHERE name in ('EMAIL_VERIFICATION_SUBJECT', 'EMAIL_VERIFICATION_REDIRECT_URL', 'EMAIL_VERIFICATION_BODY')")
	if err != nil {
		return EmailTemplateDetails{}, err
	}

	for rows.Next() {
		var name string
		var value sql.NullString
		if err := rows.Scan(&name, &value); err != nil {
			return EmailTemplateDetails{}, err
		}

		// if the value is null then return an error
		if !value.Valid {
			return EmailTemplateDetails{}, fmt.Errorf("value for %s is null", name)
		}

		switch name {
		case "EMAIL_VERIFICATION_SUBJECT":
			emailTemplate.Subject = value.String
		case "EMAIL_VERIFICATION_REDIRECT_URL":
			emailTemplate.RedirectURL = value.String
		case "EMAIL_VERIFICATION_BODY":
			emailTemplate.BodyTemplate = value.String
		}

		if err := rows.Err(); err != nil {
			return EmailTemplateDetails{}, err
		}
	}

	return emailTemplate, nil
}

func GetPasswordResetEmailDetails(db *sql.DB) (EmailTemplateDetails, error) {
	emailTemplate := EmailTemplateDetails{}

	rows, err := db.Query("SELECT name, value FROM settings WHERE name in ('PASSWORD_RESET_SUBJECT', 'PASSWORD_RESET_REDIRECT_URL', 'PASSWORD_RESET_BODY')")
	if err != nil {
		return EmailTemplateDetails{}, err
	}

	for rows.Next() {
		var name string
		var value sql.NullString
		if err := rows.Scan(&name, &value); err != nil {
			return EmailTemplateDetails{}, err
		}

		// if the value is null then return an error
		if !value.Valid {
			return EmailTemplateDetails{}, fmt.Errorf("value for %s is null", name)
		}

		switch name {
		case "PASSWORD_RESET_SUBJECT":
			emailTemplate.Subject = value.String
		case "PASSWORD_RESET_REDIRECT_URL":
			emailTemplate.RedirectURL = value.String
		case "PASSWORD_RESET_BODY":
			emailTemplate.BodyTemplate = value.String
		}

		if err := rows.Err(); err != nil {
			return EmailTemplateDetails{}, err
		}
	}

	return emailTemplate, nil
}

func GetMFAEmailDetails(db *sql.DB) (EmailTemplateDetails, error) {
	emailTemplate := EmailTemplateDetails{}

	rows, err := db.Query("SELECT name, value FROM settings WHERE name in ('MFA_VERIFICATION_SUBJECT', 'MFA_VERIFICATION_REDIRECT_URL', 'MFA_VERIFICATION_BODY')")
	if err != nil {
		return EmailTemplateDetails{}, err
	}

	for rows.Next() {
		var name string
		var value sql.NullString
		if err := rows.Scan(&name, &value); err != nil {
			return EmailTemplateDetails{}, err
		}

		// if the value is null then return an error
		if !value.Valid {
			return EmailTemplateDetails{}, fmt.Errorf("value for %s is null", name)
		}

		switch name {
		case "MFA_VERIFICATION_SUBJECT":
			emailTemplate.Subject = value.String
		case "MFA_VERIFICATION_REDIRECT_URL":
			emailTemplate.RedirectURL = value.String
		case "MFA_VERIFICATION_BODY":
			emailTemplate.BodyTemplate = value.String
		}

		if err := rows.Err(); err != nil {
			return EmailTemplateDetails{}, err
		}
	}

	return emailTemplate, nil
}

func ParseBodyTemplate(details EmailTemplateDetails, templateName string) (string, error) {
	// check if the body or the redirectURL are empty
	if details.BodyTemplate == "" || details.RedirectURL == "" {
		return "", fmt.Errorf("body template or redirect URL is empty")
	}

	tmpl, err := template.New(templateName).Parse(details.BodyTemplate)
	if err != nil {
		return "", err
	}

	// execute the template abd write the output to the body
	var emailBodyBuffer bytes.Buffer
	err = tmpl.Execute(&emailBodyBuffer, struct {
		RedirectURL string
	}{
		RedirectURL: details.RedirectURL,
	})
	if err != nil {
		return "", err
	}

	return emailBodyBuffer.String(), nil
}
