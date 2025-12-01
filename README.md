# Mood Analyst - Spotify Integration

A sentiment-inclined agent that analyzes your mood and recommends music from Spotify.

## Setup Instructions

### 1. Get Spotify API Credentials

1. Go to [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)
2. Log in or create a Spotify account
3. Create a new application to get your `Client ID` and `Client Secret`
4. Accept the terms and create the app

### 2. Configure Environment Variables

1. Copy `.env.example` to `.env`:
   ```bash
   cp .env.example .env
   ```

2. Add your Spotify credentials:
   ```
   SPOTIFY_CLIENT_ID=your_client_id_here
   SPOTIFY_CLIENT_SECRET=your_client_secret_here
   ```

3. (Optional) Add Teneo Agent SDK credentials if using NFT features:
   ```
   PRIVATE_KEY=your_private_key_here
   NFT_TOKEN_ID=your_nft_token_id_here
   OWNER_ADDRESS=your_owner_address_here
   ```

### 3. Install Dependencies

```bash
go mod download
go mod tidy
```

### 4. Run the Application

```bash
go run main.go
```

## Usage

The agent supports the `mood_analyzer` command:

```
mood_analyzer [mood description]
```

### Examples

```
mood_analyzer I feel happy and energetic
mood_analyzer I'm feeling sad and lonely
mood_analyzer I want to relax and chill
mood_analyzer I'm focused and working
mood_analyzer I'm in a romantic mood
```

## How It Works

1. **Mood Detection**: The agent analyzes your mood description and identifies the primary mood
2. **Profile Generation**: Based on the detected mood, it creates a music profile with Spotify audio features:
   - Energy level
   - Danceability
   - Valence (musical positiveness)
   - Acousticness
3. **Music Search**: Uses the Spotify API to search for tracks matching your mood
4. **Recommendations**: Returns the top 5 recommended tracks with links to play them on Spotify

## Supported Moods

- **Happy**: Upbeat, joyful, excited music
- **Sad**: Emotional, melancholic, soulful tracks
- **Relaxed**: Calm, peaceful, ambient music
- **Energetic**: High-energy, powerful, intense tracks
- **Romantic**: Love songs, passionate music
- **Focused**: Concentration-friendly music

## Project Structure

```
mood_analyst/
├── main.go                 # Main application and agent handler
├── go.mod                  # Go module file
├── .env.example            # Example environment variables
├── spotify/
│   └── client.go          # Spotify API client
└── mood/
    └── analyzer.go        # Mood analysis and music recommendations
```

## Features

- 🎵 Spotify API integration for real music recommendations
- 🧠 Mood detection from natural language descriptions
- 🎯 Smart audio feature matching (energy, danceability, valence, etc.)
- 🔗 Direct Spotify links for each recommendation
- 📱 Works with Teneo Agent SDK for multi-agent orchestration

## Architecture

### Spotify Client (`spotify/client.go`)

- `NewClient()`: Initialize the Spotify API client
- `Authenticate()`: Get access token using Client Credentials Flow
- `SearchTracks()`: Search for songs based on query
- `GetRecommendations()`: Get recommendations based on seed tracks and mood parameters
- `LoadFromEnv()`: Load credentials from environment variables

### Mood Analyzer (`mood/analyzer.go`)

- `AnalyzeMood()`: Analyze mood description and return mood profile
- `GetMoodParameters()`: Generate Spotify audio feature targets
- `FormatTrackRecommendation()`: Format track data for display

## Error Handling

The agent gracefully handles:
- Missing Spotify credentials
- Authentication failures
- API rate limits
- No results found scenarios

## Security Notes

- **Never commit `.env`** with real credentials
- Use the provided `.env.example` as a template
- Keep your Spotify credentials secure
- The Client Credentials Flow is used (server-to-server, no user login required)

## Future Enhancements

- Playlist creation from recommendations
- User preference learning
- Mood history tracking
- Multi-language support
- Real-time mood tracking with wearables

## Dependencies

- Go 1.25+
- github.com/TeneoProtocolAI/teneo-agent-sdk
- github.com/joho/godotenv (for .env file loading)
- Standard library HTTP and JSON packages

## License

See LICENSE file in the teneo-agent-sdk for licensing information.
