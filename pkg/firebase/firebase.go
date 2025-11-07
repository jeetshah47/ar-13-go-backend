package firebase

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/ar-13-go-backend/internal/config"
	"google.golang.org/api/option"
)

var (
	firebaseApp     *firebase.App
	firestoreClient *firestore.Client
	authClient      *auth.Client
)

// InitializeFirebase initializes Firebase Admin SDK
func InitializeFirebase(cfg *config.Config) (*firebase.App, error) {
	// Replace escaped newlines in private key
	// privateKey := strings.ReplaceAll(cfg.FirebasePrivateKey, "\\n", "\n")
	privateKey := cfg.FirebasePrivateKey

	opt := option.WithCredentialsJSON([]byte(fmt.Sprintf(`{
		"type": "service_account",
		"project_id": "%s",
		"private_key_id": "",
		"private_key": "%s",
		"client_email": "%s",
		"client_id": "",
		"auth_uri": "https://accounts.google.com/o/oauth2/auth",
		"token_uri": "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url": ""
	}`, cfg.FirebaseProjectID, privateKey, cfg.FirebaseClientEmail)))

	app, err := firebase.NewApp(context.Background(), &firebase.Config{
		ProjectID: cfg.FirebaseProjectID,
	}, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing Firebase app: %w", err)
	}

	firebaseApp = app

	// Initialize Firestore
	firestoreClient, err = app.Firestore(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error initializing Firestore: %w", err)
	}

	// Initialize Auth
	authClient, err = app.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error initializing Auth: %w", err)
	}

	return app, nil
}

// GetFirebaseApp returns the Firebase app instance
func GetFirebaseApp() *firebase.App {
	return firebaseApp
}

// GetFirestoreClient returns the Firestore client
func GetFirestoreClient() *firestore.Client {
	return firestoreClient
}

// GetFirebaseClient returns the Auth client wrapper
func GetFirebaseClient() *AuthClient {
	return &AuthClient{client: authClient}
}

// GetCollection returns a Firestore collection reference
func GetCollection(name string) *firestore.CollectionRef {
	return firestoreClient.Collection(name)
}
