### Getting started

This repo contains a simple web server and a React frontend. The frontend connects to the server via a websocket and displays reviews that need to be approved. Reviews are submitted to the server via the `/api/review` endpoint, and their status is polled from the `/api/review/status` endpoint.

Start the webserver and frontend with docker compose:
```bash
docker compose up
```

Then send a review to the interface via the `/api/review` endpoint:
```bash
curl -X POST http://localhost:8080/api/review \
     -H "Content-Type: application/json" \
     -d @example/payload.json
```

Check the status of the review with the `/api/review/status` endpoint:
```bash
curl http://localhost:8080/api/review/status?id=request-123
```

(replacing `request-123` with the ID of the review you submitted)

### Environment variables

The following environment variables can be set:

- `API_BASE_URL`: The base URL for the Approvals API.
- `WEBSOCKET_BASE_URL`: The base URL for the Approvals websocket.
- `OPENAI_API_KEY`: Optional. API key to use for the OpenAI API. This is used for the language model explanations of the agents actions.
