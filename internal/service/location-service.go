package service

import (
	"grocerics-backend/internal/repository"

	"gorm.io/gorm"
)

type LocationResolver struct {
	user    *repository.UserRepository
	city    *repository.CityRepository
	address *repository.AddressRepository
}

func NewLocationResolver(db *gorm.DB) *LocationResolver {
	return &LocationResolver{
		user:    repository.NewUserRepository(db),
		city:    repository.NewCityRepository(db),
		address: repository.NewAddressRepository(db),
	}
}

func (r *LocationResolver) Resolve(userID string) (cityID, pincode string, err error) {
	u, err := r.user.FindByID(userID)
	if err != nil {
		return "", "", err
	}
	if u != nil && u.CurrentCityID != nil {
		cityID = *u.CurrentCityID
	} else {
		cities, cErr := r.city.ListEnabled()
		if cErr != nil {
			return "", "", cErr
		}
		if len(cities) > 0 {
			cityID = cities[0].ID
		}
	}

	addrs, err := r.address.ListByUser(userID)
	if err != nil {
		return cityID, "", err
	}
	if len(addrs) > 0 {
		pincode = addrs[0].Pincode // ListByUser orders default first
	}
	return cityID, pincode, nil
}
