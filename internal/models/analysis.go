package models

import "time"

type PaginatedRequest struct {
	Limit  int32 `query:"limit"`
	Offset int32 `query:"offset"`
}

type PaginatedResponse[T any] struct {
	Data   []T   `json:"data"`
	Total  int64 `json:"total"`
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type Stats struct {
	Min    float32 `json:"min"`
	Max    float32 `json:"max"`
	Avg    float32 `json:"avg"`
	Median float32 `json:"median"`
}

type Analysis struct {
	ID           int32     `json:"id"`
	DateTime     time.Time `json:"date_time"`
	Product      string    `json:"product"`
	ColorRhs     string    `json:"color_rhs"`
	IDUser       string    `json:"id_user"`
	TelegramLink string    `json:"telegram_link"`
	Text         string    `json:"text"`
	FileSource   string    `json:"file_source"`
	ScaleMmPixel float64   `json:"scale_mm_pixel"`
	Mass         float64   `json:"mass"`
	Area         float64   `json:"area"`
	R            Stats     `json:"r"`
	G            Stats     `json:"g"`
	B            Stats     `json:"b"`
	H            Stats     `json:"h"`
	S            Stats     `json:"s"`
	V            Stats     `json:"v"`
	LabL         Stats     `json:"lab_l"`
	LabA         Stats     `json:"lab_a"`
	LabB         Stats     `json:"lab_b"`
	W            Stats     `json:"w"`
	L            Stats     `json:"l"`
	T            Stats     `json:"t"`
	FileOutput   string    `json:"file_output"`
	IDAnalysis   int64     `json:"id_analysis"`
	Objects      []Object  `json:"objects"`
}

type Object struct {
	ID         int32   `json:"id"`
	IdAnalysis int64   `json:"id_analysis"`
	File       string  `json:"file"`
	Class      string  `json:"class"`
	Geometry   string  `json:"geometry"`
	MH         float64 `json:"m_h"`
	MS         float64 `json:"m_s"`
	MV         float64 `json:"m_v"`
	MR         float64 `json:"m_r"`
	MG         float64 `json:"m_g"`
	MB         float64 `json:"m_b"`
	LAvg       float64 `json:"l_avg"`
	WAvg       float64 `json:"w_avg"`
	BrtAvg     float64 `json:"brt_avg"`
	RAvg       float64 `json:"r_avg"`
	GAvg       float64 `json:"g_avg"`
	BAvg       float64 `json:"b_avg"`
	HAvg       float64 `json:"h_avg"`
	SAvg       float64 `json:"s_avg"`
	VAvg       float64 `json:"v_avg"`
	H          float64 `json:"h"`
	S          float64 `json:"s"`
	V          float64 `json:"v"`
	HM         float64 `json:"h_m"`
	SM         float64 `json:"s_m"`
	VM         float64 `json:"v_m"`
	RM         float64 `json:"r_m"`
	GM         float64 `json:"g_m"`
	BM         float64 `json:"b_m"`
	BrtM       float64 `json:"brt_m"`
	WM         float64 `json:"w_m"`
	LM         float64 `json:"l_m"`
	L          float64 `json:"l"`
	W          float64 `json:"w"`
	LW         float64 `json:"l_w"`
	Pr         float64 `json:"pr"`
	Sq         float64 `json:"sq"`
	Brt        float64 `json:"brt"`
	R          float64 `json:"r"`
	G          float64 `json:"g"`
	B          float64 `json:"b"`
	Solid      float64 `json:"solid"`
	MinH       float64 `json:"min_h"`
	MinS       float64 `json:"min_s"`
	MinV       float64 `json:"min_v"`
	MaxH       float64 `json:"max_h"`
	MaxS       float64 `json:"max_s"`
	MaxV       float64 `json:"max_v"`
	Entropy    float64 `json:"entropy"`
	IDImage    int64   `json:"id_image"`
	ColorRhs   string  `json:"color_rhs"`
	SqSqcrl    float64 `json:"sq_sqcrl"`
	Hu1        float64 `json:"hu1"`
	Hu2        float64 `json:"hu2"`
	Hu3        float64 `json:"hu3"`
	Hu4        float64 `json:"hu4"`
	Hu5        float64 `json:"hu5"`
	Hu6        float64 `json:"hu6"`
}

type GetAnalysesPaginatedRequest struct {
	PaginatedRequest
	Product   string `query:"product"`
	ID        string `query:"id"`
	SortBy    string `query:"sort_by" validate:"omitempty,oneof=date_time id product"`
	SortOrder string `query:"sort_order" validate:"omitempty,oneof=asc desc"`
}
