package models

type VideoCompositionResponseSchema struct {
	Schema               string   `json:"$schema"`
	ID                   string   `json:"$id"`
	Title                string   `json:"title"`
	Description          string   `json:"description"`
	Type                 string   `json:"type"`
	Required             []string `json:"required"`
	AdditionalProperties bool     `json:"additionalProperties"`
	Properties           struct {
		Metadata struct {
			Type                 string   `json:"type"`
			Required             []string `json:"required"`
			AdditionalProperties bool     `json:"additionalProperties"`
			Properties           struct {
				TotalDuration struct {
					Type        string `json:"type"`
					Minimum     int    `json:"minimum"`
					Maximum     int    `json:"maximum"`
					Description string `json:"description"`
				} `json:"totalDuration"`
				AspectRatio struct {
					Type        string   `json:"type"`
					Enum        []string `json:"enum"`
					Description string   `json:"description"`
				} `json:"aspectRatio"`
				Fps struct {
					Type        string `json:"type"`
					Enum        []int  `json:"enum"`
					Description string `json:"description"`
				} `json:"fps"`
				Resolution struct {
					Type        string `json:"type"`
					MinItems    int    `json:"minItems"`
					MaxItems    int    `json:"maxItems"`
					Description string `json:"description"`
					Items       []struct {
						Type string `json:"type"`
					} `json:"items"`
					OneOf []struct {
						Description string  `json:"description"`
						Enum        [][]int `json:"enum"`
					} `json:"oneOf"`
				} `json:"resolution"`
			} `json:"properties"`
		} `json:"metadata"`
		Theme struct {
			Type                 string   `json:"type"`
			Required             []string `json:"required"`
			AdditionalProperties bool     `json:"additionalProperties"`
			Properties           struct {
				Style struct {
					Type string   `json:"type"`
					Enum []string `json:"enum"`
				} `json:"style"`
				Mood struct {
					Type string   `json:"type"`
					Enum []string `json:"enum"`
				} `json:"mood"`
				ColorPalette struct {
					Type                 string `json:"type"`
					AdditionalProperties bool   `json:"additionalProperties"`
					Properties           struct {
						Grading struct {
							Type        string   `json:"type"`
							Enum        []string `json:"enum"`
							Description string   `json:"description"`
						} `json:"grading"`
					} `json:"properties"`
				} `json:"colorPalette"`
				Typography struct {
					Type                 string `json:"type"`
					AdditionalProperties bool   `json:"additionalProperties"`
					Properties           struct {
						FontFamily struct {
							Type        string `json:"type"`
							Description string `json:"description"`
						} `json:"fontFamily"`
						TextStyle struct {
							Type string   `json:"type"`
							Enum []string `json:"enum"`
						} `json:"textStyle"`
					} `json:"properties"`
				} `json:"typography"`
			} `json:"properties"`
		} `json:"theme"`
		Narrative struct {
			Type                 string   `json:"type"`
			Required             []string `json:"required"`
			AdditionalProperties bool     `json:"additionalProperties"`
			Properties           struct {
				Hook struct {
					Type        string `json:"type"`
					MinLength   int    `json:"minLength"`
					MaxLength   int    `json:"maxLength"`
					Description string `json:"description"`
				} `json:"hook"`
				Story struct {
					Type  string `json:"type"`
					Items struct {
						Type      string `json:"type"`
						MinLength int    `json:"minLength"`
						MaxLength int    `json:"maxLength"`
					} `json:"items"`
					MinItems    int    `json:"minItems"`
					MaxItems    int    `json:"maxItems"`
					Description string `json:"description"`
				} `json:"story"`
				Cta struct {
					Type        string `json:"type"`
					MinLength   int    `json:"minLength"`
					MaxLength   int    `json:"maxLength"`
					Description string `json:"description"`
				} `json:"cta"`
				Tone struct {
					Type string   `json:"type"`
					Enum []string `json:"enum"`
				} `json:"tone"`
			} `json:"properties"`
		} `json:"narrative"`
		Timeline struct {
			Type     string `json:"type"`
			MinItems int    `json:"minItems"`
			Items    struct {
				Type                 string   `json:"type"`
				Required             []string `json:"required"`
				AdditionalProperties bool     `json:"additionalProperties"`
				Properties           struct {
					ID struct {
						Type      string `json:"type"`
						MinLength int    `json:"minLength"`
					} `json:"id"`
					StartTime struct {
						Type    string `json:"type"`
						Minimum int    `json:"minimum"`
					} `json:"startTime"`
					Duration struct {
						Type    string  `json:"type"`
						Minimum float64 `json:"minimum"`
					} `json:"duration"`
					Type struct {
						Type string   `json:"type"`
						Enum []string `json:"enum"`
					} `json:"type"`
					Content struct {
						AnyOf []struct {
							Ref string `json:"$ref"`
						} `json:"anyOf"`
					} `json:"content"`
				} `json:"properties"`
			} `json:"items"`
		} `json:"timeline"`
		Audio struct {
			Type                 string   `json:"type"`
			Required             []string `json:"required"`
			AdditionalProperties bool     `json:"additionalProperties"`
			Properties           struct {
				Narration struct {
					Type                 string   `json:"type"`
					Required             []string `json:"required"`
					AdditionalProperties bool     `json:"additionalProperties"`
					Properties           struct {
						Script struct {
							Type     string `json:"type"`
							MinItems int    `json:"minItems"`
							Items    struct {
								Type                 string   `json:"type"`
								Required             []string `json:"required"`
								AdditionalProperties bool     `json:"additionalProperties"`
								Properties           struct {
									Text struct {
										Type      string `json:"type"`
										MinLength int    `json:"minLength"`
									} `json:"text"`
									Timing struct {
										Type                 string   `json:"type"`
										Required             []string `json:"required"`
										AdditionalProperties bool     `json:"additionalProperties"`
										Properties           struct {
											Start struct {
												Type    string `json:"type"`
												Minimum int    `json:"minimum"`
											} `json:"start"`
											End struct {
												Type    string `json:"type"`
												Minimum int    `json:"minimum"`
											} `json:"end"`
										} `json:"properties"`
									} `json:"timing"`
									Emphasis struct {
										Type    string   `json:"type"`
										Enum    []string `json:"enum"`
										Default string   `json:"default"`
									} `json:"emphasis"`
								} `json:"properties"`
							} `json:"items"`
						} `json:"script"`
						Voice struct {
							Type                 string `json:"type"`
							AdditionalProperties bool   `json:"additionalProperties"`
							Properties           struct {
								VoiceID struct {
									Type        string `json:"type"`
									Description string `json:"description"`
									MinLength   int    `json:"minLength"`
								} `json:"voiceId"`
								Speed struct {
									Type    string  `json:"type"`
									Minimum float64 `json:"minimum"`
									Maximum float64 `json:"maximum"`
									Default float64 `json:"default"`
								} `json:"speed"`
								Pitch struct {
									Type    string  `json:"type"`
									Minimum float64 `json:"minimum"`
									Maximum float64 `json:"maximum"`
									Default float64 `json:"default"`
								} `json:"pitch"`
								Stability struct {
									Type    string  `json:"type"`
									Minimum int     `json:"minimum"`
									Maximum int     `json:"maximum"`
									Default float64 `json:"default"`
								} `json:"stability"`
							} `json:"properties"`
						} `json:"voice"`
					} `json:"properties"`
				} `json:"narration"`
				Music struct {
					Type                 string   `json:"type"`
					Required             []string `json:"required"`
					AdditionalProperties bool     `json:"additionalProperties"`
					Properties           struct {
						Enabled struct {
							Type string `json:"type"`
						} `json:"enabled"`
						TrackID struct {
							Type        string `json:"type"`
							Description string `json:"description"`
						} `json:"trackId"`
						Genre struct {
							Type string   `json:"type"`
							Enum []string `json:"enum"`
						} `json:"genre"`
						Mood struct {
							Type string   `json:"type"`
							Enum []string `json:"enum"`
						} `json:"mood"`
						Volume struct {
							Type    string  `json:"type"`
							Minimum int     `json:"minimum"`
							Maximum int     `json:"maximum"`
							Default float64 `json:"default"`
						} `json:"volume"`
					} `json:"properties"`
					Description string `json:"description"`
				} `json:"music"`
			} `json:"properties"`
		} `json:"audio"`
	} `json:"properties"`
	Defs struct {
		ImageSegment struct {
			Type                 string   `json:"type"`
			Required             []string `json:"required"`
			AdditionalProperties bool     `json:"additionalProperties"`
			Properties           struct {
				ImageIndex struct {
					Type    string `json:"type"`
					Minimum int    `json:"minimum"`
				} `json:"imageIndex"`
				Animation struct {
					Type                 string `json:"type"`
					AdditionalProperties bool   `json:"additionalProperties"`
					Properties           struct {
						Type struct {
							Type        string   `json:"type"`
							Enum        []string `json:"enum"`
							Description string   `json:"description"`
						} `json:"type"`
						Intensity struct {
							Type        string  `json:"type"`
							Minimum     float64 `json:"minimum"`
							Maximum     float64 `json:"maximum"`
							Default     float64 `json:"default"`
							Description string  `json:"description"`
						} `json:"intensity"`
					} `json:"properties"`
				} `json:"animation"`
				Crop struct {
					Type                 string `json:"type"`
					AdditionalProperties bool   `json:"additionalProperties"`
					Description          string `json:"description"`
					Properties           struct {
						X struct {
							Type        string `json:"type"`
							Minimum     int    `json:"minimum"`
							Maximum     int    `json:"maximum"`
							Description string `json:"description"`
						} `json:"x"`
						Y struct {
							Type        string `json:"type"`
							Minimum     int    `json:"minimum"`
							Maximum     int    `json:"maximum"`
							Description string `json:"description"`
						} `json:"y"`
						Width struct {
							Type        string  `json:"type"`
							Minimum     float64 `json:"minimum"`
							Maximum     int     `json:"maximum"`
							Description string  `json:"description"`
						} `json:"width"`
						Height struct {
							Type        string  `json:"type"`
							Minimum     float64 `json:"minimum"`
							Maximum     int     `json:"maximum"`
							Description string  `json:"description"`
						} `json:"height"`
					} `json:"properties"`
				} `json:"crop"`
			} `json:"properties"`
		} `json:"ImageSegment"`
		TextOverlay struct {
			Type                 string `json:"type"`
			AdditionalProperties bool   `json:"additionalProperties"`
			Properties           struct {
				Text struct {
					Type      string `json:"type"`
					MinLength int    `json:"minLength"`
				} `json:"text"`
				Position struct {
					Type        string   `json:"type"`
					Enum        []string `json:"enum"`
					Default     string   `json:"default"`
					Description string   `json:"description"`
				} `json:"position"`
				Style struct {
					Type                 string `json:"type"`
					AdditionalProperties bool   `json:"additionalProperties"`
					Properties           struct {
						FontSize struct {
							Type    string `json:"type"`
							Minimum int    `json:"minimum"`
							Maximum int    `json:"maximum"`
							Default int    `json:"default"`
						} `json:"fontSize"`
						Color struct {
							Type    string `json:"type"`
							Pattern string `json:"pattern"`
						} `json:"color"`
						BackgroundColor struct {
							Type    string `json:"type"`
							Pattern string `json:"pattern"`
						} `json:"backgroundColor"`
						Animation struct {
							Type    string   `json:"type"`
							Enum    []string `json:"enum"`
							Default string   `json:"default"`
						} `json:"animation"`
					} `json:"properties"`
					Required []string `json:"required"`
				} `json:"style"`
			} `json:"properties"`
		} `json:"TextOverlay"`
		Transition struct {
			Type                 string `json:"type"`
			AdditionalProperties bool   `json:"additionalProperties"`
			Properties           struct {
				Effect struct {
					Type    string   `json:"type"`
					Enum    []string `json:"enum"`
					Default string   `json:"default"`
				} `json:"effect"`
				Easing struct {
					Type    string   `json:"type"`
					Enum    []string `json:"enum"`
					Default string   `json:"default"`
				} `json:"easing"`
			} `json:"properties"`
		} `json:"Transition"`
	} `json:"$defs"`
}
