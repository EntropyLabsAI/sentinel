# Asteroid Web
This is a React app that provides a UI for interacting with the Asteroid server. 

Currently, the frontend supports:
- Reviewing requests made to the Human-in-the-loop Supervisor 
- Viewing reviews that the LLM Supervisor has made 

You should run this in tandem with the [Approvals server](../README.md) by running `docker compose up` from the root directory.

### Development
```bash
npm install
npm run dev
```

We use [Orval](https://orval.dev/quick-start) to generate the API client and server from the OpenAPI spec. If you update the OpenAPI spec in `../openapi.yaml`, you should regenerate the API client and server with `oapi-codegen` so that the endpoints are available in the frontend.

```bash
npx orval --input ../server/openapi.yaml --output ./src/types.ts
```

### Build
```bash
npm run build
```


