package chart

// DesiredState defines the desired install state of a helm chart.
type DesiredState struct {
	ReleaseNamespace string
	ReleaseName      string
	ChartName        string
	MinVersion       string
	ExactVersion     string
	Values           map[string]any
	SkipInstall      bool
}
