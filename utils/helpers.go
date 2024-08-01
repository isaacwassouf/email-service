package utils

import (
	"errors"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pbCrypto "github.com/isaacwassou/email-service/protobufs/cryptography_service"
	pb "github.com/isaacwassou/email-service/protobufs/email_management_service"
)

type EmailTemplateDBFields struct {
	Subject     string
	Body        string
	RedirectURL string
}

func GetEmailTemplateDBFields(emailType pb.EmailType) (fields EmailTemplateDBFields, err error) {
	switch emailType {
	case pb.EmailType_EMAIL_VERIFICATION:
		fields.Subject = "EMAIL_VERIFICATION_SUBJECT"
		fields.Body = "EMAIL_VERIFICATION_BODY"
		fields.RedirectURL = "EMAIL_VERIFICATION_REDIRECT_URL"
		break

	case pb.EmailType_PASSWORD_RESET:
		fields.Subject = "PASSWORD_RESET_SUBJECT"
		fields.Body = "PASSWORD_RESET_BODY"
		fields.RedirectURL = "PASSWORD_RESET_REDIRECT_URL"
		break

	case pb.EmailType_MFA:
		fields.Subject = "MFA_VERIFICATION_SUBJECT"
		fields.Body = "MFA_VERIFICATION_BODY"
		fields.RedirectURL = "MFA_VERIFICATION_REDIRECT_URL"
		break

	default:
		return fields, errors.New("Invalid email type")
	}

	return fields, nil
}

func NewCryptoServiceClient() (pbCrypto.CryptographyManagerClient, error) {
	// get host and port from the environment
	host, found := os.LookupEnv("CRYPTOGRAPHY_SERVICE_HOST")
	if !found {
		host = "localhost"
	}

	port, found := os.LookupEnv("CRYPTOGRAPHY_SERVICE_PORT")
	if !found {
		port = "8094"
	}

	connectionURI := host + ":" + port

	// Create a connection to the email service
	conn, err := grpc.Dial(
		connectionURI,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return pbCrypto.NewCryptographyManagerClient(conn), nil
}

func GetGoEnv() string {
	environment, found := os.LookupEnv("GO_ENV")
	if !found {
		return "development"
	}

	return environment
}
