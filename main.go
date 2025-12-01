package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"example/mood_analyst/mood"
	"example/mood_analyst/spotify"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
	"github.com/joho/godotenv"
)

type MoodalystAgent struct {
	spotifyClient *spotify.Client
	moodAnalyzer  *mood.MoodAnalyzer
}

func (a *MoodalystAgent) ProcessTask(ctx context.Context, task string) (string, error) {
	log.Printf("Processing task: %s", task)

	// Clean up the task input
	task = strings.TrimSpace(task)
	task = strings.TrimPrefix(task, "/")
	taskLower := strings.ToLower(task)

	// Split into command and arguments
	parts := strings.Fields(taskLower)
	if len(parts) == 0 {
		return "No command provided. Available commands: mood_analyzer", nil
	}

	command := parts[0]
	args := parts[1:]

	// Route to appropriate command handler
	switch command {
	case "mood_analyzer":
		if len(args) == 0 {
			return "Please describe your mood. Example: 'mood_analyzer I feel happy and energetic'", nil
		}

		moodDescription := strings.Join(args, " ")
		return a.recommendMusic(ctx, moodDescription)

	default:
		return fmt.Sprintf("Unknown command '%s'. Available commands: mood_analyzer", command), nil
	}
}

// recommendMusic analyzes the mood and recommends music from Spotify
func (a *MoodalystAgent) recommendMusic(_ context.Context, moodDescription string) (string, error) {
	// Analyze the mood
	moodProfile := a.moodAnalyzer.AnalyzeMood(moodDescription)
	log.Printf("Detected mood: %s", moodProfile.Mood)

	// Search for tracks matching the mood
	query := moodProfile.SearchQueryTerms
	if query == "" {
		query = moodDescription
	}

	tracks, err := a.spotifyClient.SearchTracks(query, 5)
	if err != nil {
		log.Printf("Error searching tracks: %v", err)
		return fmt.Sprintf("I detected your mood as '%s', but I couldn't fetch recommendations right now. Try again later!", moodProfile.Mood), nil
	}

	if len(tracks) == 0 {
		return fmt.Sprintf("I understand you're feeling %s, but I couldn't find any matching songs right now.", moodProfile.Mood), nil
	}

	// Get 15 additional recommendations to make a total of 20 tracks
	var seedTrackIDs []string
	for _, t := range tracks {
		if t.ID != "" {
			seedTrackIDs = append(seedTrackIDs, t.ID)
			log.Printf("Adding seed track ID: %s (Name: %s)", t.ID, t.Name)
		}
	}

	// Spotify allows max 5 seeds. We use the tracks we found as seeds.
	// If we have fewer than 5 tracks, we can fill up with genres.
	var seedGenres []string
	if len(seedTrackIDs) < 5 {
		remaining := 5 - len(seedTrackIDs)
		if len(moodProfile.SuggestedGenres) > 0 {
			if len(moodProfile.SuggestedGenres) > remaining {
				seedGenres = moodProfile.SuggestedGenres[:remaining]
			} else {
				seedGenres = moodProfile.SuggestedGenres
			}
		}
	}

	moodParams := map[string]interface{}{
		"target_energy":       moodProfile.Energy,
		"target_danceability": moodProfile.Danceability,
		"target_valence":      moodProfile.Valence,
		"target_acousticness": moodProfile.Acousticness,
	}

	log.Printf("Fetching 15 additional recommendations using %d seed tracks and %d genres", len(seedTrackIDs), len(seedGenres))
	recs, err := a.spotifyClient.GetRecommendations(seedTrackIDs, seedGenres, moodParams, 15)
	if err == nil {
		log.Printf("Successfully got %d recommendations, appending to %d existing tracks", len(recs), len(tracks))
		tracks = append(tracks, recs...)
		log.Printf("Total tracks now: %d", len(tracks))
	} else {
		log.Printf("Failed to get recommendations: %v", err)
		// Fallback: Do additional searches with different mood keywords
		log.Printf("Trying fallback: searching for more tracks with mood keywords")
		fallbackQuery := fmt.Sprintf("%s %s", query, moodProfile.Mood)
		moreTracks, searchErr := a.spotifyClient.SearchTracks(fallbackQuery, 15)
		if searchErr == nil && len(moreTracks) > 0 {
			log.Printf("Fallback successful: found %d additional tracks", len(moreTracks))
			tracks = append(tracks, moreTracks...)
		} else {
			log.Printf("Fallback also failed: %v", searchErr)
		}
	}

	// Build response with recommendations
	response := fmt.Sprintf("Based on your mood (%s), here are some song recommendations:\n\n", moodProfile.Mood)
	var trackURIs []string

	log.Printf("Building response with %d total tracks", len(tracks))
	for i, track := range tracks {
		artistName := "Unknown"
		if len(track.Artists) > 0 {
			artistName = track.Artists[0].Name
		}
		recommendation := mood.FormatTrackRecommendation(track.Name, artistName, track.ExternalURLs.Spotify)
		response += fmt.Sprintf("%d. %s\n", i+1, recommendation)
		if track.URI != "" {
			trackURIs = append(trackURIs, track.URI)
		}
	}

	// Try to create a playlist if we have user access
	user, err := a.spotifyClient.GetCurrentUser()
	if err == nil && user != nil {
		playlistName := fmt.Sprintf("Mood Analyst: %s Vibes", strings.Title(moodProfile.Mood))
		description := fmt.Sprintf("A playlist curated for your %s mood.", moodProfile.Mood)

		playlist, err := a.spotifyClient.CreatePlaylist(user.ID, playlistName, description)
		if err == nil {
			log.Printf("Created playlist, adding %d tracks", len(trackURIs))
			err = a.spotifyClient.AddTracksToPlaylist(playlist.ID, trackURIs)
			if err == nil {
				response += fmt.Sprintf("\n✨ I've also created a playlist for you: %s\n", playlist.ExternalURLs.Spotify)
			} else {
				log.Printf("Failed to add tracks to playlist: %v", err)
			}
		} else {
			log.Printf("Failed to create playlist: %v", err)
		}
	} else {

		log.Printf("Skipping playlist creation (user not authenticated or scope missing): %v", err)
	}

	return response, nil
}

func main() {
	godotenv.Load()
	config := agent.DefaultConfig()

	config.Name = "mood analyst"
	config.Description = "A sentiment inclined agent, that recommends music based on your mood(a user describes their mood, then the agent recommends a music by pulling the music from spotify)"
	config.Capabilities = []string{"mood_analysis", "sentiment_analysis", "emotion_analysis"}
	config.PrivateKey = os.Getenv("PRIVATE_KEY")
	config.NFTTokenID = os.Getenv("NFT_TOKEN_ID")
	config.OwnerAddress = os.Getenv("OWNER_ADDRESS")

	// Initialize Spotify client
	spotifyClient, err := spotify.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize Spotify client: %v", err)
	}

	// Authenticate with Spotify
	if err := spotifyClient.Authenticate(); err != nil {
		log.Fatalf("Failed to authenticate with Spotify: %v", err)
	}

	log.Println("Successfully authenticated with Spotify")

	moodAnalyzer := &mood.MoodAnalyzer{}

	enhancedAgent, err := agent.NewEnhancedAgent(&agent.EnhancedAgentConfig{
		Config: config,
		AgentHandler: &MoodalystAgent{
			spotifyClient: spotifyClient,
			moodAnalyzer:  moodAnalyzer,
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Starting mood analyst...")
	enhancedAgent.Run()
}
