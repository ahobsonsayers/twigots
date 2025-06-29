package twigots

import (
	"strings"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
)

func FuzzySubstringMatch(
	substring, target string,
	gapPenalty float64,
) float64 {
	// Split strings up into words
	subWords := strings.Fields(substring)
	targetWords := strings.Fields(target)

	numSubWords := len(subWords)
	numTargetWords := len(targetWords)

	// If both have no words, exit early
	if numSubWords == 0 && numTargetWords == 0 {
		return 1
	}

	// If only one has no words, exit early
	if numSubWords == 0 || numTargetWords == 0 {
		return 0
	}

	numRows := numSubWords + 1
	numCols := numTargetWords + 1

	matrix := make([][]float64, numRows)
	for i := range matrix {
		matrix[i] = make([]float64, numCols)
	}

	// // Initialize first column with gap penalties (cost to skip substring words)
	// for i := 1; i < numRows; i++ {
	// 	matrix[i][0] = matrix[i-1][0] - gapPenalty
	// }

	for i := 1; i < numRows; i++ {
		for j := 1; j < numCols; j++ {
			wordSim := calculateWordSimilarity(subWords[i-1], targetWords[j-1])

			matchScore := matrix[i-1][j-1] + wordSim
			deleteScore := matrix[i-1][j] - gapPenalty
			insertScore := matrix[i][j-1] - gapPenalty

			matrix[i][j] = maxUtil(0, matchScore, deleteScore, insertScore)
		}
	}

	// Only look at scores where we've consumed ALL substring words (last row)
	// But allow ending anywhere in the target
	maxScore := maxUtil(matrix[numSubWords][numTargetWords])

	avgScore := maxScore / float64(numSubWords)
	return avgScore
}

func calculateWordSimilarity(stringA, stringB string) float64 {
	return strutil.Similarity(
		stringA, stringB,
		metrics.NewLevenshtein(),
	)
}

func maxUtil(nums ...float64) float64 {
	maxNum := nums[0]
	for _, num := range nums[1:] {
		if num > maxNum {
			maxNum = num
		}
	}
	return maxNum
}
