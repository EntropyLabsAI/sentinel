# Submit a review request via curl

Curl the payload to the `/api/review` endpoint to submit a review request, updating the base URL to point to your Approvals server.

```bash
curl -X POST http://localhost:8080/api/review \
     -H "Content-Type: application/json" \
     -d @payload.json
```

