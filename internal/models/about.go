package models

type AboutResponse struct {
	Company        string   `json:"company"`
	Description    string   `json:"description"`
	Advantages     []string `json:"advantages"`
	Regions        []string `json:"regions"`
	PaymentMethods []string `json:"payment_methods"`
	CooperationURL string   `json:"cooperation_url,omitempty"`
	VacanciesURL   string   `json:"vacancies_url,omitempty"`
}
