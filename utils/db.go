package utils

import (
	"database/sql"
	"strconv"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Password string
	User     string
	Sender   string
}

func GetSMTPConfig(db *sql.DB) (SMTPConfig, error) {
	smtpConfig := SMTPConfig{}

	rows, err := db.Query("SELECT name, value FROM settings WHERE name IN ('SMTP_HOST', 'SMTP_PORT', 'SMTP_USER', 'SMTP_PASSWORD', 'SMTP_SENDER')")
	if err != nil {
		return SMTPConfig{}, err
	}

	for rows.Next() {
		var name string
		var value sql.NullString
		if err := rows.Scan(&name, &value); err != nil {
			return SMTPConfig{}, err
		}

		switch name {
		case "SMTP_HOST":
			smtpConfig.Host = value.String
		case "SMTP_PORT":
			// convert the port to an integer
			port, err := strconv.Atoi(value.String)
			if err != nil {
				return SMTPConfig{}, err
			}
			smtpConfig.Port = port
		case "SMTP_USER":
			smtpConfig.User = value.String
		case "SMTP_PASSWORD":
			smtpConfig.Password = value.String
		case "SMTP_SENDER":
			smtpConfig.Sender = value.String
		}

		if err := rows.Err(); err != nil {
			return SMTPConfig{}, err
		}
	}

	return smtpConfig, nil
}

func CheckSMTPConfig(smtpConfig SMTPConfig) bool {
	if smtpConfig.Host == "" || smtpConfig.Port == 0 || smtpConfig.User == "" || smtpConfig.Password == "" || smtpConfig.Sender == "" {
		return false
	}

	return true
}
