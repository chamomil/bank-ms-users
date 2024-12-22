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

	UserWorkplace struct {
		CompanyName    string `json:"companyName"`
		CompanyAddress string `json:"companyAddress"`
		Position       string `json:"position"`
		StartDate      int64  `json:"startDate"`
		EndDate        *int64 `json:"endDate"`
	}

	Workplace struct {
		CompanyName    string  `json:"companyName"`
		CompanyAddress string  `json:"companyAddress"`
		Position       string  `json:"position"`
		StartDate      string  `json:"startDate"`
		EndDate        *string `json:"endDate"`
	}
)
