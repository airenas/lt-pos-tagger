package service

//ResultWord is service output
type ResultWord struct {
	Type   string `json:"type"`
	String string `json:"string,omitempty"`
	Mi     string `json:"mi,omitempty"`
	Lemma  string `json:"lemma,omitempty"`
}
