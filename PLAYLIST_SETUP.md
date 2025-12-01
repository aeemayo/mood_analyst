# Setting Up Playlist Creation

To enable the Mood Analyst to create playlists on your behalf, you need to provide a **Refresh Token**. This allows the agent to access your Spotify account securely.

## Prerequisites

1.  A Spotify Developer App (which you should already have).
2.  Your `Client ID` and `Client Secret`.

## Step 1: Configure Redirect URI

1.  Go to your [Spotify Developer Dashboard](https://developer.spotify.com/dashboard).
2.  Select your app.
3.  Click "Edit Settings".
4.  Under "Redirect URIs", add: `http://localhost:8888/callback`
5.  Click "Save".

## Step 2: Get Authorization Code

Open the following URL in your browser (replace `YOUR_CLIENT_ID` with your actual Client ID):

```
https://accounts.spotify.com/authorize?client_id=e3741e80012b4d61969552bb7f997886&response_type=code&redirect_uri=https://well-xfjz.onrender.com/spotify/callback&scope=playlist-modify-public%20playlist-modify-private%20user-read-private
```

1.  Log in to Spotify if asked.
2.  Click "Agree".
3.  You will be redirected to a URL like `http://localhost:8888/callback?code=NApCCg...`
4.  Copy the code after `code=` (everything before any `&` if present).

## Step 3: Get Refresh Token

You need to exchange this code for a refresh token. You can use `curl` or a tool like Postman.

**Using curl:**

Replace `YOUR_CLIENT_ID`, `YOUR_CLIENT_SECRET`, and `THE_CODE_FROM_STEP_2` below:

```bash
curl -H "Authorization: Basic $(echo -n 'YOUR_CLIENT_ID:YOUR_CLIENT_SECRET' | base64)" -d grant_type=authorization_code -d code=THE_CODE_FROM_STEP_2 -d redirect_uri=http://localhost:8888/callback https://accounts.spotify.com/api/token
```

*Note: If you are on Windows Command Prompt, the base64 part might be tricky. You can use an online base64 encoder to encode `CLIENT_ID:CLIENT_SECRET` and use `Authorization: Basic <encoded_string>`.*

**Response:**

You will get a JSON response containing `access_token` and `refresh_token`.

## Step 4: Update .env

Add the refresh token to your `.env` file:

```
SPOTIFY_REFRESH_TOKEN=your_refresh_token_here
```

## Step 5: Restart the Agent

Restart your agent (or rebuild the Docker container). It will now automatically create playlists for your mood recommendations!









