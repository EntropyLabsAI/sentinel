### Getting started

Start the webserver 
```bash
docker compose up
```

Then send a review to the interface
```bash
curl -X POST   -H "Content-Type: application/json"   -d '{ 
    "id": "request-123",
    "context": "Sample context for testing.",
    "proposed_action": "Proposed action for testing."
  }'   http://localhost:8080/api/review
```
