package myerror

type Error struct {
	Location  string `json:"location"`
	Parameter string `json:"param"`
	Value     string `json:"value"`
	MSG       string `json:"msg"`
}
