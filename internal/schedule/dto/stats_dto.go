package dto

type DayStat struct {
	Date      string `json:"date"`
	Completed int    `json:"completed"`
	Pending   int    `json:"pending"`
}

type GetStatsResponse struct {
	Week []DayStat `json:"week"`
}
