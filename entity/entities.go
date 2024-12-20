package entity

type (
	UserPersonalData struct {
		PhoneNumber     string  `json:"phoneNumber"`
		FirstName       string  `json:"firstName"`
		LastName        string  `json:"lastName"`
		FathersName     *string `json:"fathersName"`
		DateOfBirth     string  `json:"dateOfBirth"`
		PassportId      string  `json:"passportId"`
		Address         string  `json:"address"`
		Gender          string  `json:"gender"`
		LiveInCountryId int64   `json:"liveInCountry"`
	}
)
