import React, { useEffect } from 'react';
import SwaggerUI from "swagger-ui-react"
import "swagger-ui-react/swagger-ui.css"

// The API base URL is set via an environment variable in the docker-compose.yml file
// @ts-ignore
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

const SwaggerUIComponent: React.FC = () => {
  return <SwaggerUI url={`${API_BASE_URL}/api/v1/openapi.yaml`} />;
};

export default SwaggerUIComponent;
