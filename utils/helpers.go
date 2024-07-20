package utils

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pbCrypto "github.com/isaacwassou/email-service/protobufs/cryptography_service"
)

func NewCryptoServiceClient() (pbCrypto.CryptographyManagerClient, error) {
	// Create a connection to the email service
	conn, err := grpc.Dial(
		"localhost:8094",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return pbCrypto.NewCryptographyManagerClient(conn), nil
}
