package engine

import (
	"math"
	"sort"
)

func BuildTree(examples []Example, features []FeatureMeta) *DecisionTree {
	root := buildRecursive(examples, features)
	return &DecisionTree{Root: root, Features: features}
}

func buildRecursive(examples []Example, features []FeatureMeta) *TreeNode {
	if len(examples) == 0 {
		return &TreeNode{IsLeaf: true, Label: LabelLow}
	}

	if allSameLabel(examples) {
		return &TreeNode{IsLeaf: true, Label: examples[0].Label}
	}

	if len(features) == 0 {
		return &TreeNode{IsLeaf: true, Label: majorityLabel(examples)}
	}

	if len(examples) < 5 {
		return &TreeNode{IsLeaf: true, Label: majorityLabel(examples)}
	}

	bestFeature, bestThreshold := findBestSplit(examples, features)

	left, right := splitContinuous(examples, bestFeature, bestThreshold)

	remainingFeatures := removeFeature(features, bestFeature)

	node := &TreeNode{
		Feature:     bestFeature,
		Threshold:   bestThreshold,
		LessEqual:   buildRecursive(left, remainingFeatures),
		GreaterThan: buildRecursive(right, remainingFeatures),
	}

	return node
}

func allSameLabel(examples []Example) bool {
	if len(examples) == 0 {
		return true
	}
	first := examples[0].Label
	for _, ex := range examples[1:] {
		if ex.Label != first {
			return false
		}
	}
	return true
}

func majorityLabel(examples []Example) Label {
	counts := make(map[Label]int)
	for _, ex := range examples {
		counts[ex.Label]++
	}

	var best Label
	bestCount := 0
	for label, count := range counts {
		if count > bestCount {
			best = label
			bestCount = count
		}
	}
	return best
}

func findBestSplit(examples []Example, features []FeatureMeta) (string, float64) {
	baseEntropy := entropy(examples)
	total := float64(len(examples))

	bestFeature := features[0].Name
	bestThreshold := 0.0
	bestGainRatio := -1.0

	for _, f := range features {
		if f.Type == Categorical {
			continue
		}

		threshold, gainRatio := findBestContinuousSplit(examples, f.Name, baseEntropy, total)
		if gainRatio > bestGainRatio {
			bestGainRatio = gainRatio
			bestFeature = f.Name
			bestThreshold = threshold
		}
	}

	return bestFeature, bestThreshold
}

func findBestContinuousSplit(examples []Example, feature string, baseEntropy float64, total float64) (float64, float64) {
	sort.Slice(examples, func(i, j int) bool {
		return examples[i].Features[feature] < examples[j].Features[feature]
	})

	bestGainRatio := -1.0
	bestThreshold := examples[len(examples)/2].Features[feature]

	for i := 0; i < len(examples)-1; i++ {
		if examples[i].Label == examples[i+1].Label {
			continue
		}

		threshold := (examples[i].Features[feature] + examples[i+1].Features[feature]) / 2.0

		leftCount := 0
		leftLabels := make(map[Label]int)
		rightCount := 0
		rightLabels := make(map[Label]int)

		for _, ex := range examples {
			if ex.Features[feature] <= threshold {
				leftCount++
				leftLabels[ex.Label]++
			} else {
				rightCount++
				rightLabels[ex.Label]++
			}
		}

		if leftCount == 0 || rightCount == 0 {
			continue
		}

		leftEntropy := calcEntropyFromCounts(leftLabels, leftCount)
		rightEntropy := calcEntropyFromCounts(rightLabels, rightCount)
		splitEntropy := (float64(leftCount)/total)*leftEntropy + (float64(rightCount)/total)*rightEntropy
		infoGain := baseEntropy - splitEntropy

		splitInfo := splitInfo(float64(leftCount)/total, float64(rightCount)/total)

		var gainRatio float64
		if splitInfo > 0 {
			gainRatio = infoGain / splitInfo
		} else {
			continue
		}

		if gainRatio > bestGainRatio {
			bestGainRatio = gainRatio
			bestThreshold = threshold
		}
	}

	return bestThreshold, bestGainRatio
}

func splitContinuous(examples []Example, feature string, threshold float64) ([]Example, []Example) {
	var left, right []Example
	for _, ex := range examples {
		if ex.Features[feature] <= threshold {
			left = append(left, ex)
		} else {
			right = append(right, ex)
		}
	}
	return left, right
}

func entropy(examples []Example) float64 {
	counts := make(map[Label]int)
	for _, ex := range examples {
		counts[ex.Label]++
	}
	return calcEntropyFromCounts(counts, len(examples))
}

func calcEntropyFromCounts(counts map[Label]int, total int) float64 {
	if total == 0 {
		return 0
	}
	h := 0.0
	for _, count := range counts {
		p := float64(count) / float64(total)
		if p > 0 {
			h -= p * math.Log2(p)
		}
	}
	return h
}

func splitInfo(ratios ...float64) float64 {
	si := 0.0
	for _, r := range ratios {
		if r > 0 {
			si -= r * math.Log2(r)
		}
	}
	return si
}

func removeFeature(features []FeatureMeta, name string) []FeatureMeta {
	var result []FeatureMeta
	for _, f := range features {
		if f.Name != name {
			result = append(result, f)
		}
	}
	return result
}
