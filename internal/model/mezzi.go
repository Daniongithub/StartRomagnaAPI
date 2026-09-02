package model

type VehicleInService struct {
	Number    string  `db:"matricola" json:"number"`
	PlateNum  string  `db:"targa" json:"plate_num"`
	Model     string  `db:"modello" json:"model"`
	Basin     string  `db:"provincia" json:"basin"`
	PhotoPath *string `db:"photo_path" json:"photo_path"`
}
