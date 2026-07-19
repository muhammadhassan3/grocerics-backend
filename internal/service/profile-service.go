package service

import (
	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/errs"
	"grocerics-backend/internal/repository"

	"gorm.io/gorm"
)

type ProfileService struct {
	user    *repository.UserRepository
	address *repository.AddressRepository
	pincode *repository.PincodeRepository
	city    *repository.CityRepository
	notif   *repository.NotificationPreferenceRepository
	fcm     *repository.FcmTokenRepository
}

func NewProfileService(db *gorm.DB) *ProfileService {
	return &ProfileService{
		user:    repository.NewUserRepository(db),
		address: repository.NewAddressRepository(db),
		pincode: repository.NewPincodeRepository(db),
		city:    repository.NewCityRepository(db),
		notif:   repository.NewNotificationPreferenceRepository(db),
		fcm:     repository.NewFcmTokenRepository(db),
	}
}

func (s *ProfileService) GetMe(userID string) (*dto.MeDTO, error) {
	u, err := s.user.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errs.NotFound("USER_NOT_FOUND", "user not found")
	}
	me := &dto.MeDTO{ID: u.ID, Name: u.Name, Phone: u.Phone}
	if u.CurrentCityID != nil {
		me.CurrentCityID = *u.CurrentCityID
		if c, _ := s.city.FindByID(*u.CurrentCityID); c != nil {
			me.CurrentCityName = c.Name
		}
	}
	return me, nil
}

func (s *ProfileService) UpdateMe(userID, name string) (*dto.MeDTO, error) {
	if name != "" {
		if _, err := s.user.Update(&domain.User{BaseModel: domain.BaseModel{ID: userID}, Name: name}); err != nil {
			return nil, err
		}
	}
	return s.GetMe(userID)
}

type AddressInput struct {
	Label     *string
	Line1     string
	Line2     *string
	Pincode   string
	Lat       *float64
	Lng       *float64
	IsDefault bool
}

func (s *ProfileService) ListAddresses(userID string) ([]dto.AddressDTO, error) {
	addrs, err := s.address.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.AddressDTO, 0, len(addrs))
	for _, a := range addrs {
		out = append(out, s.toAddressDTO(a))
	}
	return out, nil
}

func (s *ProfileService) CreateAddress(userID string, in AddressInput) (*dto.AddressDTO, error) {
	cityID := s.resolveCity(in.Pincode)
	a := &domain.UserAddress{
		UserID: userID, Label: in.Label, Line1: in.Line1, Line2: in.Line2,
		Pincode: in.Pincode, CityID: cityID, Lat: in.Lat, Lng: in.Lng, IsDefault: in.IsDefault,
	}
	if in.IsDefault {
		if err := s.address.UnsetDefaults(userID); err != nil {
			return nil, err
		}
	}
	created, err := s.address.Create(a)
	if err != nil {
		return nil, err
	}
	if in.IsDefault {
		if err := s.user.SetCurrentCity(userID, cityID); err != nil {
			return nil, err
		}
	}
	d := s.toAddressDTO(*created)
	return &d, nil
}

func (s *ProfileService) UpdateAddress(userID, addressID string, in AddressInput) (*dto.AddressDTO, error) {
	a, err := s.address.FindByID(addressID)
	if err != nil {
		return nil, err
	}
	if a == nil || a.UserID != userID {
		return nil, errs.NotFound("ADDRESS_NOT_FOUND", "address not found")
	}
	a.Label = in.Label
	a.Line1 = in.Line1
	a.Line2 = in.Line2
	a.Lat = in.Lat
	a.Lng = in.Lng
	a.IsDefault = in.IsDefault
	if in.Pincode != "" && in.Pincode != a.Pincode {
		a.Pincode = in.Pincode
		a.CityID = s.resolveCity(in.Pincode)
	}
	if in.IsDefault {
		if err := s.address.UnsetDefaults(userID); err != nil {
			return nil, err
		}
	}
	updated, err := s.address.Update(a)
	if err != nil {
		return nil, err
	}
	if in.IsDefault {
		if err := s.user.SetCurrentCity(userID, a.CityID); err != nil {
			return nil, err
		}
	}
	d := s.toAddressDTO(*updated)
	return &d, nil
}

func (s *ProfileService) DeleteAddress(userID, addressID string) error {
	a, err := s.address.FindByID(addressID)
	if err != nil {
		return err
	}
	if a == nil || a.UserID != userID {
		return errs.NotFound("ADDRESS_NOT_FOUND", "address not found")
	}
	return s.address.Delete(addressID)
}

func (s *ProfileService) resolveCity(pincode string) *string {
	row, err := s.pincode.FindByPincode(pincode)
	if err != nil || row == nil || !row.Serviceable {
		return nil
	}
	cid := row.CityID
	return &cid
}

func (s *ProfileService) toAddressDTO(a domain.UserAddress) dto.AddressDTO {
	d := dto.AddressDTO{ID: a.ID, Line1: a.Line1, Pincode: a.Pincode, Lat: a.Lat, Lng: a.Lng, IsDefault: a.IsDefault}
	if a.Label != nil {
		d.Label = *a.Label
	}
	if a.Line2 != nil {
		d.Line2 = *a.Line2
	}
	if a.CityID != nil {
		d.CityID = *a.CityID
		d.Serviceable = true
		if c, _ := s.city.FindByID(*a.CityID); c != nil {
			d.CityName = c.Name
		}
	}
	return d
}

func (s *ProfileService) GetNotificationPreferences(userID string) (*dto.NotificationPreferencesDTO, error) {
	p, err := s.notif.Get(userID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return &dto.NotificationPreferencesDTO{PriceAlerts: true, Promotions: true, OrderUpdates: true}, nil
	}
	return &dto.NotificationPreferencesDTO{PriceAlerts: p.PriceAlerts, Promotions: p.Promotions, OrderUpdates: p.OrderUpdates}, nil
}

func (s *ProfileService) UpdateNotificationPreferences(userID string, in dto.NotificationPreferencesDTO) (*dto.NotificationPreferencesDTO, error) {
	_, err := s.notif.Upsert(&domain.NotificationPreference{
		UserID: userID, PriceAlerts: in.PriceAlerts, Promotions: in.Promotions, OrderUpdates: in.OrderUpdates,
	})
	if err != nil {
		return nil, err
	}
	return &in, nil
}

func (s *ProfileService) RegisterFcmToken(userID, token, platform string) error {
	p := domain.DevicePlatform(platform)
	if !p.IsValid() {
		p = domain.DevicePlatformAndroid
	}
	return s.fcm.Upsert(userID, token, p)
}

func (s *ProfileService) RemoveFcmToken(token string) error { return s.fcm.Delete(token) }
