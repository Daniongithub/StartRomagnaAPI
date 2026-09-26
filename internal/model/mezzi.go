package model

type VehicleInService struct {
	Number         string  `db:"matricola" json:"number"`
	PlateNum       *string `db:"targa" json:"plate_num"`
	Model          *string `db:"modello" json:"model"`
	Basin          *string `db:"provincia" json:"basin"`
	BusPagePath    *string `db:"path_html" json:"bus_page_path"`
	BusPreviewPath *string `db:"photo_path" json:"bus_preview_path"`
}

type VehicleReport struct {
	Number string  `db:"matricola" json:"number"`
	Basin  *string `db:"provincia" json:"basin"`
}
