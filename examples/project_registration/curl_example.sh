source .env &&curl -X POST http://localhost:${APPROVAL_WEBSERVER_PORT}/api/project \
     -H "Content-Type: application/json" \
     -d @examples/project_registration/payload.json

# Example get projects
curl http://localhost:${APPROVAL_WEBSERVER_PORT}/api/project
