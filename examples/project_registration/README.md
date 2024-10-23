# Register a project via curl

You can register a project by sending a POST request to the `/api/project` endpoint.

From the root of the repo:
```bash
source .env && curl -X POST http://localhost:${APPROVAL_WEBSERVER_PORT}/api/project \
     -H "Content-Type: application/json" \
     -d @examples/project_registration/payload.json
```

You should see output like the following:

```json
{"id":"123e4567-e89b-12d3-a456-426614174000"}
```

You can then view the review at http://localhost:${APPROVAL_WEBSERVER_PORT}/api/project/123e4567-e89b-12d3-a456-426614174000
