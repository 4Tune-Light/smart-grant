package engine

type Label string

const (
	LabelLow    Label = "low"
	LabelMedium Label = "medium"
	LabelHigh   Label = "high"
)

type FeatureType int

const (
	Continuous  FeatureType = 0
	Categorical FeatureType = 1
)

type FeatureMeta struct {
	Name string
	Type FeatureType
}

type Example struct {
	Features map[string]float64
	Label    Label
}

type TreeNode struct {
	IsLeaf      bool
	Label       Label
	Feature     string
	Threshold   float64
	LessEqual   *TreeNode
	GreaterThan *TreeNode
}

type DecisionTree struct {
	Root     *TreeNode
	Features []FeatureMeta
}

func (t *DecisionTree) Classify(features map[string]float64) (Label, float64) {
	return t.classifyRecursive(t.Root, features, 1.0)
}

func (t *DecisionTree) classifyRecursive(node *TreeNode, features map[string]float64, confidence float64) (Label, float64) {
	if node.IsLeaf {
		return node.Label, confidence
	}

	val, ok := features[node.Feature]
	if !ok {
		return t.majorityLabel(), confidence * 0.5
	}

	if val <= node.Threshold {
		return t.classifyRecursive(node.LessEqual, features, confidence)
	}
	return t.classifyRecursive(node.GreaterThan, features, confidence)
}

func (t *DecisionTree) majorityLabel() Label {
	return LabelLow
}
