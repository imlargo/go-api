package evaluator

// SimilarityResult representa el resultado de comparación entre videos
type SimilarityResult struct {
	OriginalVideo   string
	ComparedVideo   string
	Similarity      float64
	IsOriginal      bool
	MatchedSegments int
	TotalSegments   int
	HashDistance    int
}
