package engine

func GenerateSeedData() []Example {
	return []Example{
		// LOW RISK — amount wajar, jarang/baru, dokumen lengkap
		{Features: map[string]float64{"nominal_amount": 100000000, "funding_frequency_30d": 0, "document_completeness": 1.0}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 200000000, "funding_frequency_30d": 1, "document_completeness": 0.9}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 500000000, "funding_frequency_30d": 0, "document_completeness": 1.0}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 750000000, "funding_frequency_30d": 1, "document_completeness": 0.95}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 300000000, "funding_frequency_30d": 0, "document_completeness": 0.8}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 150000000, "funding_frequency_30d": 0, "document_completeness": 1.0}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 450000000, "funding_frequency_30d": 1, "document_completeness": 0.9}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 80000000, "funding_frequency_30d": 0, "document_completeness": 1.0}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 600000000, "funding_frequency_30d": 1, "document_completeness": 0.85}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 250000000, "funding_frequency_30d": 2, "document_completeness": 1.0}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 50000000, "funding_frequency_30d": 0, "document_completeness": 1.0}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 900000000, "funding_frequency_30d": 1, "document_completeness": 0.9}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 350000000, "funding_frequency_30d": 0, "document_completeness": 0.95}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 550000000, "funding_frequency_30d": 2, "document_completeness": 0.8}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 180000000, "funding_frequency_30d": 0, "document_completeness": 1.0}, Label: LabelLow},
		{Features: map[string]float64{"nominal_amount": 720000000, "funding_frequency_30d": 1, "document_completeness": 1.0}, Label: LabelLow},

		// MEDIUM RISK — amount besar, frekuensi sedang, dokumen agak kurang
		{Features: map[string]float64{"nominal_amount": 800000000, "funding_frequency_30d": 2, "document_completeness": 0.6}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 900000000, "funding_frequency_30d": 3, "document_completeness": 0.7}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 1000000000, "funding_frequency_30d": 2, "document_completeness": 0.5}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 1100000000, "funding_frequency_30d": 3, "document_completeness": 0.65}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 1200000000, "funding_frequency_30d": 1, "document_completeness": 0.5}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 300000000, "funding_frequency_30d": 4, "document_completeness": 0.6}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 500000000, "funding_frequency_30d": 4, "document_completeness": 0.55}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 700000000, "funding_frequency_30d": 3, "document_completeness": 0.5}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 850000000, "funding_frequency_30d": 2, "document_completeness": 0.4}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 1000000000, "funding_frequency_30d": 0, "document_completeness": 0.3}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 400000000, "funding_frequency_30d": 5, "document_completeness": 0.65}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 650000000, "funding_frequency_30d": 3, "document_completeness": 0.45}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 950000000, "funding_frequency_30d": 2, "document_completeness": 0.35}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 1300000000, "funding_frequency_30d": 1, "document_completeness": 0.4}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 200000000, "funding_frequency_30d": 6, "document_completeness": 0.6}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 600000000, "funding_frequency_30d": 5, "document_completeness": 0.5}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 780000000, "funding_frequency_30d": 3, "document_completeness": 0.55}, Label: LabelMedium},
		{Features: map[string]float64{"nominal_amount": 1050000000, "funding_frequency_30d": 2, "document_completeness": 0.45}, Label: LabelMedium},

		// HIGH RISK — amount besar, frekuensi tinggi, dokumen kurang
		{Features: map[string]float64{"nominal_amount": 2000000000, "funding_frequency_30d": 3, "document_completeness": 0.3}, Label: LabelHigh},
		{Features: map[string]float64{"nominal_amount": 3000000000, "funding_frequency_30d": 4, "document_completeness": 0.2}, Label: LabelHigh},
		{Features: map[string]float64{"nominal_amount": 5000000000, "funding_frequency_30d": 5, "document_completeness": 0.1}, Label: LabelHigh},
		{Features: map[string]float64{"nominal_amount": 2500000000, "funding_frequency_30d": 4, "document_completeness": 0.25}, Label: LabelHigh},
		{Features: map[string]float64{"nominal_amount": 1500000000, "funding_frequency_30d": 6, "document_completeness": 0.15}, Label: LabelHigh},
		{Features: map[string]float64{"nominal_amount": 4000000000, "funding_frequency_30d": 5, "document_completeness": 0.3}, Label: LabelHigh},
		{Features: map[string]float64{"nominal_amount": 3500000000, "funding_frequency_30d": 3, "document_completeness": 0.2}, Label: LabelHigh},
		{Features: map[string]float64{"nominal_amount": 1800000000, "funding_frequency_30d": 7, "document_completeness": 0.1}, Label: LabelHigh},
		{Features: map[string]float64{"nominal_amount": 4500000000, "funding_frequency_30d": 2, "document_completeness": 0.05}, Label: LabelHigh},
		{Features: map[string]float64{"nominal_amount": 2200000000, "funding_frequency_30d": 4, "document_completeness": 0.2}, Label: LabelHigh},
		{Features: map[string]float64{"nominal_amount": 2800000000, "funding_frequency_30d": 5, "document_completeness": 0.15}, Label: LabelHigh},
		{Features: map[string]float64{"nominal_amount": 1200000000, "funding_frequency_30d": 6, "document_completeness": 0.1}, Label: LabelHigh},
		{Features: map[string]float64{"nominal_amount": 3200000000, "funding_frequency_30d": 4, "document_completeness": 0.25}, Label: LabelHigh},
		{Features: map[string]float64{"nominal_amount": 1000000000, "funding_frequency_30d": 8, "document_completeness": 0.2}, Label: LabelHigh},
		{Features: map[string]float64{"nominal_amount": 5000000000, "funding_frequency_30d": 6, "document_completeness": 0.3}, Label: LabelHigh},
		{Features: map[string]float64{"nominal_amount": 1700000000, "funding_frequency_30d": 5, "document_completeness": 0.1}, Label: LabelHigh},
	}
}

var RiskFeatures = []FeatureMeta{
	{Name: "nominal_amount", Type: Continuous},
	{Name: "funding_frequency_30d", Type: Continuous},
	{Name: "document_completeness", Type: Continuous},
}

const ModelVersion = "c4.5-v2"
