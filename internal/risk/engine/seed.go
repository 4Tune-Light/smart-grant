package engine

func GenerateSeedData() []Example {
	return []Example{
		{Features: map[string]float64{"nominal_amount": 100000000, "funding_frequency_30d": 1, "document_completeness": 0.9}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 200000000, "funding_frequency_30d": 2, "document_completeness": 0.8}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 500000000, "funding_frequency_30d": 1, "document_completeness": 1.0}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 800000000, "funding_frequency_30d": 3, "document_completeness": 0.7}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 900000000, "funding_frequency_30d": 2, "document_completeness": 0.6}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 1000000000, "funding_frequency_30d": 4, "document_completeness": 0.5}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 2000000000, "funding_frequency_30d": 3, "document_completeness": 0.3}, Label: LabelHigh},
		{Features: map[string]float64{"nominal_amount": 5000000000, "funding_frequency_30d": 5, "document_completeness": 0.2}, Label: LabelHigh},
		{Features: map[string]float64{"nominal_amount": 1500000000, "funding_frequency_30d": 6, "document_completeness": 0.1}, Label: LabelHigh},
		{Features: map[string]float64{"nominal_amount": 300000000, "funding_frequency_30d": 1, "document_completeness": 0.4}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 750000000, "funding_frequency_30d": 2, "document_completeness": 0.95}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 3000000000, "funding_frequency_30d": 4, "document_completeness": 0.6}, Label: LabelHigh},
		{Features: map[string]float64{"nominal_amount": 50000000, "funding_frequency_30d": 0, "document_completeness": 1.0}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 1200000000, "funding_frequency_30d": 3, "document_completeness": 0.5}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 4000000000, "funding_frequency_30d": 5, "document_completeness": 0.25}, Label: LabelHigh},
	}
}

var RiskFeatures = []FeatureMeta{
	{Name: "nominal_amount", Type: Continuous},
	{Name: "funding_frequency_30d", Type: Continuous},
	{Name: "document_completeness", Type: Continuous},
}
