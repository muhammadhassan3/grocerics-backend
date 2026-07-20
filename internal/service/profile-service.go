package service

import (
	"strings"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/errs"
	"grocerics-backend/internal/repository"

	"gorm.io/gorm"
)

type ProfileService struct {
	db      *gorm.DB
	user    *repository.UserRepository
	address *repository.AddressRepository
	city    *repository.CityRepository
	notif   *repository.NotificationPreferenceRepository
	fcm     *repository.FcmTokenRepository
}

func NewProfileService(db *gorm.DB) *ProfileService {
	return &ProfileService{
		db:      db,
		user:    repository.NewUserRepository(db),
		address: repository.NewAddressRepository(db),
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
	me := &dto.MeDTO{ID: u.ID, Name: u.Name, Phone: u.Phone, Onboarded: u.CurrentCityID != nil}
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
	City      string // device-geocoded city; resolved to an enabled city
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
	cityID, err := s.matchEnabledCity(in.City)
	if err != nil {
		return nil, err
	}
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
	if in.IsDefault && cityID != nil {
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
	a.Pincode = in.Pincode
	a.IsDefault = in.IsDefault
	cityID, err := s.matchEnabledCity(in.City)
	if err != nil {
		return nil, err
	}
	a.CityID = cityID
	if in.IsDefault {
		if err := s.address.UnsetDefaults(userID); err != nil {
			return nil, err
		}
	}
	updated, err := s.address.Update(a)
	if err != nil {
		return nil, err
	}
	if in.IsDefault && a.CityID != nil {
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

// cityAliases maps common reverse-geocoder names to our enabled-city names —
// device geocoders return "Bengaluru"/"New Delhi"/"Bombay" where our rows say
// "Bangalore"/"Delhi"/"Mumbai".
// ponytail: stopgap name-matching. Real fix is a geospatial radius match on the
// lat/lng onboarding already sends (city coords are seeded) — swap when revisited.
var cityAliases = map[string]string{
	"bengaluru": "bangalore",
	"new delhi": "delhi",
	"delhi ncr": "delhi",
	"bombay":    "mumbai",
	"amdavad":   "ahmedabad",
}

func normalizeCity(s string) string {
	s = strings.Join(strings.Fields(strings.ToLower(s)), " ")
	if a, ok := cityAliases[s]; ok {
		return a
	}
	return s
}

func (s *ProfileService) matchEnabledCity(city string) (*string, error) {
	norm := normalizeCity(city)
	if norm == "" {
		return nil, nil
	}
	cands := []string{norm}
	for _, w := range strings.Fields(norm) {
		cands = append(cands, normalizeCity(w))
	}
	cities, err := s.city.ListEnabled()
	if err != nil {
		return nil, err
	}
	for _, c := range cities {
		for _, w := range cands {
			if strings.EqualFold(c.Name, w) || strings.EqualFold(c.Slug, w) {
				id := c.ID
				return &id, nil
			}
		}
	}
	return nil, nil
}

func (s *ProfileService) Onboard(userID, name string, in AddressInput) (*dto.OnboardingResponse, error) {
	if strings.TrimSpace(in.City) == "" {
		return nil, errs.BadRequest("LOCATION_REQUIRED", "enable location to continue")
	}
	cityRef, err := s.matchEnabledCity(in.City)
	if err != nil {
		return nil, err
	}
	if cityRef == nil {
		return nil, errs.BadRequest("CITY_NOT_SERVICEABLE", "we don't deliver to "+in.City+" yet")
	}
	cityID := *cityRef
	in.IsDefault = true

	var addr *domain.UserAddress
	err = s.db.Transaction(func(tx *gorm.DB) error {
		ur := repository.NewUserRepository(tx)
		ar := repository.NewAddressRepository(tx)
		u, err := ur.FindByID(userID)
		if err != nil {
			return err
		}
		if u == nil {
			return errs.NotFound("USER_NOT_FOUND", "user not found")
		}
		u.Name = name
		if _, err := ur.Update(u); err != nil {
			return err
		}
		if err := ar.UnsetDefaults(userID); err != nil {
			return err
		}
		addr, err = ar.Create(&domain.UserAddress{
			UserID: userID, Label: in.Label, Line1: in.Line1, Line2: in.Line2,
			Pincode: in.Pincode, CityID: &cityID, Lat: in.Lat, Lng: in.Lng, IsDefault: true,
		})
		if err != nil {
			return err
		}
		return ur.SetCurrentCity(userID, &cityID)
	})
	if err != nil {
		return nil, err
	}

	me, err := s.GetMe(userID)
	if err != nil {
		return nil, err
	}
	d := s.toAddressDTO(*addr)
	return &dto.OnboardingResponse{User: *me, Address: d}, nil
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
		return &dto.NotificationPreferencesDTO{Muted: false, Promotions: true, Deals: true}, nil
	}
	return &dto.NotificationPreferencesDTO{Muted: p.Muted, Promotions: p.Promotions, Deals: p.Deals}, nil
}

func (s *ProfileService) UpdateNotificationPreferences(userID string, in dto.NotificationPreferencesDTO) (*dto.NotificationPreferencesDTO, error) {
	_, err := s.notif.Upsert(&domain.NotificationPreference{
		UserID: userID, Muted: in.Muted, Promotions: in.Promotions, Deals: in.Deals,
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
