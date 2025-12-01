package mood

import (
	"fmt"
	"strings"
)

// MoodAnalyzer analyzes user mood and determines music preferences
type MoodAnalyzer struct{}

// MoodProfile represents user mood characteristics
type MoodProfile struct {
	Mood             string
	Energy           float32
	Danceability     float32
	Valence          float32
	Acousticness     float32
	SuggestedGenres  []string
	SearchQueryTerms string
}

// AnalyzeMood analyzes mood description and returns mood profile
func (ma *MoodAnalyzer) AnalyzeMood(moodDescription string) MoodProfile {
	description := strings.ToLower(moodDescription)

	profile := MoodProfile{
		Mood:            "neutral",
		Energy:          0.5,
		Danceability:    0.5,
		Valence:         0.5,
		Acousticness:    0.5,
		SuggestedGenres: []string{},
	}

	// Detect happy/positive moods
	if containsAny(description, []string{"happy", "joyful", "excited", "energetic", "upbeat", "great", "fantastic"}) {
		profile.Mood = "happy"
		profile.Energy = 0.8
		profile.Danceability = 0.7
		profile.Valence = 0.8
		profile.Acousticness = 0.3
		profile.SuggestedGenres = []string{"pop", "dance", "electronic", "funk"}
		profile.SearchQueryTerms = "happy upbeat energetic"
	}

	// Detect sad/melancholic moods
	if containsAny(description, []string{"sad", "down", "depressed", "lonely", "blue", "heartbroken", "melancholy"}) {
		profile.Mood = "sad"
		profile.Energy = 0.3
		profile.Danceability = 0.2
		profile.Valence = 0.2
		profile.Acousticness = 0.7
		profile.SuggestedGenres = []string{"indie", "folk", "soul", "acoustic"}
		profile.SearchQueryTerms = "sad emotional soulful"
	}

	// Detect relaxed/calm moods
	if containsAny(description, []string{"calm", "relaxed", "chill", "peaceful", "serene", "tranquil", "zen"}) {
		profile.Mood = "relaxed"
		profile.Energy = 0.2
		profile.Danceability = 0.3
		profile.Valence = 0.5
		profile.Acousticness = 0.8
		profile.SuggestedGenres = []string{"ambient", "lo-fi", "jazz", "acoustic"}
		profile.SearchQueryTerms = "relaxing chill ambient"
	}

	// Detect energetic/pumped moods
	if containsAny(description, []string{"pumped", "energetic", "motivated", "fired up", "adrenaline"}) {
		profile.Mood = "energetic"
		profile.Energy = 0.9
		profile.Danceability = 0.8
		profile.Valence = 0.7
		profile.Acousticness = 0.1
		profile.SuggestedGenres = []string{"hip-hop", "electronic", "rock", "metal"}
		profile.SearchQueryTerms = "energetic powerful intense"
	}

	// Detect romantic/loving moods
	if containsAny(description, []string{"romantic", "in love", "loved", "affectionate", "passionate"}) {
		profile.Mood = "romantic"
		profile.Energy = 0.4
		profile.Danceability = 0.5
		profile.Valence = 0.7
		profile.Acousticness = 0.6
		profile.SuggestedGenres = []string{"soul", "r&b", "indie", "acoustic pop"}
		profile.SearchQueryTerms = "romantic love passionate"
	}

	// Detect focus/study moods
	if containsAny(description, []string{"focused", "studying", "concentrating", "working", "productive"}) {
		profile.Mood = "focused"
		profile.Energy = 0.5
		profile.Danceability = 0.3
		profile.Valence = 0.5
		profile.Acousticness = 0.5
		profile.SuggestedGenres = []string{"lo-fi", "classical", "ambient", "instrumental"}
		profile.SearchQueryTerms = "focus study concentration"
	}

	return profile
}

// GetMoodParameters returns Spotify API parameters for mood
func (ma *MoodAnalyzer) GetMoodParameters(profile MoodProfile) map[string]interface{} {
	return map[string]interface{}{
		"target_energy":       profile.Energy,
		"target_danceability": profile.Danceability,
		"target_valence":      profile.Valence,
		"target_acousticness": profile.Acousticness,
	}
}

// containsAny checks if string contains any of the given substrings
func containsAny(text string, terms []string) bool {
	for _, term := range terms {
		if strings.Contains(text, term) {
			return true
		}
	}
	return false
}

// FormatTrackRecommendation formats a track into a recommendation string
func FormatTrackRecommendation(trackName, artistName, spotifyURL string) string {
	return fmt.Sprintf("🎵 %s by %s\n   🔗 %s", trackName, artistName, spotifyURL)
}
