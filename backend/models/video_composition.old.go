package models

type OldVideoCompositionResponse struct {
	ThemeSpec         ThemeSpec `json:"theme_spec"`
	BeatSheet         BeatSheet `json:"beat_sheet"`
	ImageDescriptions []string  `json:"image_descriptions"`
}

type ThemeSpec struct {
}

type BeatSheet struct {
}
