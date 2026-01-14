# 🍼 Draw a Memory

A web application where users can upload photos to create album memories using AI, especially related to the first months of their children.

## Features

- 📸 Upload up to 10 photos at once (max 5MB each)
- 🖼️ Photo gallery with responsive grid layout
- 🤖 AI-powered story generation (coming soon)
- 👨‍👩‍👧 Family member tagging (coming soon)
- 📖 Memory book creation (coming soon)

## Tech Stack

### Client
- React 19
- TypeScript
- Vite

### Server
- Go 1.21+
- Standard library HTTP server

### AI (Coming Soon)
- Google Gemini

## Project Structure

```
draw_a_memory/
├── client/                 # React frontend
│   ├── src/
│   │   ├── api/           # API client functions
│   │   ├── components/    # React components
│   │   └── types/         # TypeScript types
│   └── package.json
├── server/                 # Go backend
│   ├── main.go            # Server entry point
│   ├── uploads/           # Uploaded photos (gitignored)
│   └── go.mod
└── README.md
```

## Getting Started

### Prerequisites

- [Node.js](https://nodejs.org/) (v18 or higher)
- [Go](https://golang.org/) (v1.21 or higher)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/YOUR_USERNAME/draw_a_memory.git
   cd draw_a_memory
   ```

2. Install client dependencies:
   ```bash
   cd client
   npm install
   ```

3. Install server dependencies:
   ```bash
   cd server
   go mod tidy
   ```

### Running the Application

1. Start the server (terminal 1):
   ```bash
   cd server
   go run main.go
   ```
   Server runs on http://localhost:8080

2. Start the client (terminal 2):
   ```bash
   cd client
   npm run dev
   ```
   Client runs on http://localhost:5173

3. Open http://localhost:5173 in your browser

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/photos/upload` | Upload photos (multipart/form-data) |
| GET | `/api/photos` | List all uploaded photos |
| GET | `/uploads/{filename}` | Serve uploaded photo |

## License

MIT


UPDATES:

- those clusters will be getting updated and visible to the user as it progress
- Then user can discard unwanted photos
- By cluster, group similar photos so it's easy for the user to discard photos
- User can also discard a whole cluster of photos

- Fix: the draft and the final version should be identical
- allow to add more photos in the cluster to fill spaces
- Allow to select between different layout templates to arrange the description text.

REMEMBER: the main idea is that the user can easily organize their memories in a really mess amount of photos