# Submit a review request via curl

Curl the payload to the `/api/review` endpoint to submit a review request, updating the base URL to point to your Approvals server.

From the root of the repo:
```bash
source .env && curl -X POST http://localhost:${APPROVAL_WEBSERVER_PORT}/api/review/human \
     -H "Content-Type: application/json" \
     -d @examples/curl_example/payload.json
```

You should see output like the following:

```json
{"review_id":"123e4567-e89b-12d3-a456-426614174000"}
```

You can then view the review at http://localhost:${APPROVAL_WEBSERVER_PORT}/api/review/123e4567-e89b-12d3-a456-426614174000
