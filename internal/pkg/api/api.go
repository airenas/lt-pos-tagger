package api

//SegmenterResult is segmentation response
type SegmenterResult struct {
	Seg [][]int `json:"seg"`
	S   [][]int `json:"s"`
	P   [][]int `json:"p"`
}

//TaggerResult is tagger response
type TaggerResult struct {
	Msd  [][][]string `json:"msd"`
	Stem []string     `json:"stem"`
}
