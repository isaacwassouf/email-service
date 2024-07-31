package utils

import (
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pbCrypto "github.com/isaacwassou/email-service/protobufs/cryptography_service"
)

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
