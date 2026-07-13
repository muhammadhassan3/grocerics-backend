package repository

import firebase "firebase.google.com/go"

type FirebaseMessagingRepository struct {
	firebase *firebase.App
	userRepo *UserRepository
	
}

func NewFirebaseMessagingRepository(firebaseApp *firebase.App) *FirebaseMessagingRepository {
	return &FirebaseMessagingRepository{
		firebase: firebaseApp,
	}
}

func 
