package model

type RouteInfoRaw struct {
	Version      string              `json:"version"`
	Areas        map[string][]Linea  `json:"areas"`
	Excluded     map[string][]string `json:"excluded"`
	Hero         Hero                `json:"hero"`
	Booklets     []Booklet           `json:"booklets"`
	LastUpdated  string              `json:"lastUpdated"`
	DataBase     string              `json:"dataBase"`
	AssetVersion string              `json:"assetVersion"`
	InfoBusAjax  string              `json:"infobusAjax"`
}

type Linea struct {
	Label     string `json:"label"`
	Title     string `json:"title"`
	Pdf       string `json:"pdf"`
	Route     string `json:"route"`
	InfoBusId int    `json:"infoBusId"`
}

type Hero struct {
	Eyebrow string `json:"eyebrow"`
	Title   string `json:"title"`
	Text    string `json:"text"`
	Image   string `json:"image"`
}

type Booklet struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Title   string `json:"title"`
	Class   string `json:"class"`
	Url     string `json:"url"`
	Updated string `json:"updated"`
}
