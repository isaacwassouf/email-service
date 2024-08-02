package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gopkg.in/gomail.v2"

	"github.com/isaacwassou/email-service/database"
	pbCrypto "github.com/isaacwassou/email-service/protobufs/cryptography_service"
	pb "github.com/isaacwassou/email-service/protobufs/email_management_service"
	"github.com/isaacwassou/email-service/utils"
)

type EmailManagerService struct {
	pb.UnimplementedEmailManagerServer
	cryptoServiceClient pbCrypto.CryptographyManagerClient
	emailServiceDB      *database.EmailServiceDB
}

func (s *EmailManagerService) SendVerifyEmailEmail(ctx context.Context, in *pb.SendEmailRequest) (*pb.SendEmailResponse, error) {
	// get the SMTP configuration from the database
	smtpConfig, err := utils.GetSMTPConfig(s.emailServiceDB.Db)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// check if the SMTP configuration is valid
	if !utils.CheckSMTPConfig(smtpConfig) {
		return nil, status.Error(codes.FailedPrecondition, "SMTP configuration is not set!")
	}

	// decrypt the SMTP Password
	decryptedPassword, err := s.cryptoServiceClient.Decrypt(ctx, &pbCrypto.DecryptRequest{Ciphertext: smtpConfig.Password})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// get the email template details from the database
	emailTemplate, err := utils.GetVerifyEmailDetails(s.emailServiceDB.Db)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// parse the email template body
	emailBody, err := utils.ParseBodyTemplate(emailTemplate, "verify-email")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// create new message
	m := gomail.NewMessage()
	// set the email message headers
	m.SetHeader("From", smtpConfig.Sender)
	m.SetHeader("To", in.To)
	m.SetHeader("Subject", emailTemplate.Subject)
	m.SetBody("text/html", emailBody)

	// create a new dialer
	dialer := gomail.NewDialer(
		smtpConfig.Host,
		smtpConfig.Port,
		smtpConfig.User,
		decryptedPassword.Plaintext,
	)

	// open a connection to the SMTP server and send the email
	if err := dialer.DialAndSend(m); err != nil {
		log.Printf("Failed to send an email to %s with error %s", in.To, err)
		return &pb.SendEmailResponse{Message: "Failed to send email!"}, nil
	}

	return &pb.SendEmailResponse{Message: "Sent an email successfully!"}, nil
}

func (s *EmailManagerService) SendPasswordResetEmail(ctx context.Context, in *pb.SendEmailRequest) (*pb.SendEmailResponse, error) {
	// get the SMTP configuration from the database
	smtpConfig, err := utils.GetSMTPConfig(s.emailServiceDB.Db)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// check if the SMTP configuration is valid
	if !utils.CheckSMTPConfig(smtpConfig) {
		return nil, status.Error(codes.FailedPrecondition, "SMTP configuration is not set!")
	}

	// decrypt the SMTP Password
	decryptedPassword, err := s.cryptoServiceClient.Decrypt(ctx, &pbCrypto.DecryptRequest{Ciphertext: smtpConfig.Password})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// get the email template details from the database
	emailTemplate, err := utils.GetPasswordResetEmailDetails(s.emailServiceDB.Db)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// add the password reset token to the redirect URL
	emailTemplate.RedirectURL = fmt.Sprintf("%s?code=%s", emailTemplate.RedirectURL, in.Token)

	// parse the email template body
	emailBody, err := utils.ParseBodyTemplate(emailTemplate, "password-reset")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// create new message
	m := gomail.NewMessage()
	// set the email message headers
	m.SetHeader("From", smtpConfig.Sender)
	m.SetHeader("To", in.To)
	m.SetHeader("Subject", emailTemplate.Subject)
	m.SetBody("text/html", emailBody)

	// create a new dialer
	dialer := gomail.NewDialer(
		smtpConfig.Host,
		smtpConfig.Port,
		smtpConfig.User,
		decryptedPassword.Plaintext,
	)

	// open a connection to the SMTP server and send the email
	if err := dialer.DialAndSend(m); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SendEmailResponse{Message: "Sent an email successfully!"}, nil
}

// SetSMTPCredentials sets the SMTP credentials in the database
func (s *EmailManagerService) SetSMTPCredentials(ctx context.Context, in *pb.SetSMTPCredentialsRequest) (*pb.SetSMTPCredentialsResponse, error) {
	tx, err := s.emailServiceDB.Db.Begin()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = tx.Exec("UPDATE settings SET value = ? WHERE name = 'SMTP_HOST'", in.Host)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = tx.Exec("UPDATE settings SET value = ? WHERE name = 'SMTP_PORT'", in.Port)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = tx.Exec("UPDATE settings SET value = ? WHERE name = 'SMTP_USER'", in.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Encrypt the password before storing it in the database
	encryptedPassword, err := s.cryptoServiceClient.Encrypt(ctx, &pbCrypto.EncryptRequest{Plaintext: in.Password})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	_, err = tx.Exec("UPDATE settings SET value = ? WHERE name = 'SMTP_PASSWORD'", encryptedPassword.Ciphertext)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = tx.Exec("UPDATE settings SET value = ? WHERE name = 'SMTP_SENDER'", in.Sender)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = tx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SetSMTPCredentialsResponse{Message: "SMTP credentials set successfully!"}, nil
}

func (s *EmailManagerService) SetEmailVerificationTemplate(ctx context.Context, in *pb.SetEmailTemplateRequest) (*pb.SetEmailTemplateResponse, error) {
	tx, err := s.emailServiceDB.Db.Begin()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = tx.Exec("UPDATE settings SET value = ? WHERE name = 'EMAIL_VERIFICATION_SUBJECT'", in.Subject)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = tx.Exec("UPDATE settings SET value = ? WHERE name = 'EMAIL_VERIFICATION_BODY'", in.Body)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = tx.Exec("UPDATE settings SET value = ? WHERE name = 'EMAIL_VERIFICATION_REDIRECT_URL'", in.RedirectUrl)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = tx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SetEmailTemplateResponse{Message: "Email template set successfully!"}, nil
}

func (s *EmailManagerService) GetSMTPCredentials(ctx context.Context, in *emptypb.Empty) (*pb.SetSMTPCredentialsRequest, error) {
	var host, user, sender sql.NullString
	var port sql.NullInt32

	err := s.emailServiceDB.Db.QueryRow("SELECT value FROM settings WHERE name = 'SMTP_HOST'").Scan(&host)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.emailServiceDB.Db.QueryRow("SELECT value FROM settings WHERE name = 'SMTP_PORT'").Scan(&port)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.emailServiceDB.Db.QueryRow("SELECT value FROM settings WHERE name = 'SMTP_USER'").Scan(&user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.emailServiceDB.Db.QueryRow("SELECT value FROM settings WHERE name = 'SMTP_SENDER'").Scan(&sender)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SetSMTPCredentialsRequest{
		Host:     host.String,
		Port:     port.Int32,
		Username: user.String,
		Sender:   sender.String,
	}, nil
}

func (s *EmailManagerService) SetEmailTemplate(ctx context.Context, in *pb.SetEmailTemplateRequest) (*pb.SetEmailTemplateResponse, error) {
	emailTemplateFields, err := utils.GetEmailTemplateDBFields(in.EmailType)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	tx, err := s.emailServiceDB.Db.Begin()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	updateSubjectQuery := fmt.Sprintf("UPDATE settings SET value = ? WHERE name = '%s'", emailTemplateFields.Subject)
	_, err = tx.Exec(updateSubjectQuery, in.Subject)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	updateBodyQuery := fmt.Sprintf("UPDATE settings SET value = ? WHERE name = '%s'", emailTemplateFields.Body)
	_, err = tx.Exec(updateBodyQuery, in.Body)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	updateRedirectURLQuery := fmt.Sprintf("UPDATE settings SET value = ? WHERE name = '%s'", emailTemplateFields.RedirectURL)
	_, err = tx.Exec(updateRedirectURLQuery, in.RedirectUrl)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = tx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SetEmailTemplateResponse{Message: "Email template set successfully!"}, nil
}

func (s *EmailManagerService) GetEmailTemplate(ctx context.Context, in *pb.GetEmailTemaplateRequest) (*pb.GetEmailTemaplateResponse, error) {
	emailTemplateFields, err := utils.GetEmailTemplateDBFields(in.EmailType)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var subject, body, redirectURL sql.NullString

	err = s.emailServiceDB.Db.QueryRow("SELECT value FROM settings WHERE name = ?", emailTemplateFields.Subject).Scan(&subject)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.emailServiceDB.Db.QueryRow("SELECT value FROM settings WHERE name = ?", emailTemplateFields.Body).Scan(&body)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.emailServiceDB.Db.QueryRow("SELECT value FROM settings WHERE name = ?", emailTemplateFields.RedirectURL).Scan(&redirectURL)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetEmailTemaplateResponse{
		Subject:     subject.String,
		Body:        body.String,
		RedirectUrl: redirectURL.String,
	}, nil
}

func main() {
	// Get the environment i.e. development or production
	environment := utils.GetGoEnv()

	// Load the .env file if the environment is development
	if environment == "development" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Failed to load the .env file: ", err)
		}
	}

	// Create a new cryptoServiceClient
	cryptoServiceClient, err := utils.NewCryptoServiceClient()
	if err != nil {
		log.Fatalf("failed to create a new CryptoServiceClient: %v", err)
	}

	// Create a new schemaManagementServiceDB
	emailServiceDB, err := database.NewEmailServiceDB()
	if err != nil {
		log.Fatalf("failed to create a new SchemaManagementServiceDB: %v", err)
	}
	// ping the database
	err = emailServiceDB.Db.Ping()
	if err != nil {
		log.Fatalf("failed to ping the database: %v", err)
	}

	// create a listener on TCP port 8080
	ls, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Failed to listen: ", err)
	}

	// Close the listener when the application exits
	defer ls.Close()

	fmt.Println("Server started on port 8080")

	s := grpc.NewServer()
	pb.RegisterEmailManagerServer(s, &EmailManagerService{emailServiceDB: emailServiceDB, cryptoServiceClient: cryptoServiceClient})

	if err := s.Serve(ls); err != nil {
		log.Fatal("Failed to serve the gRPC server: ", err)
	}
}
