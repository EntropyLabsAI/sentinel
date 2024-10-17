import React, { useEffect } from 'react';
import SwaggerUI from "swagger-ui-react"
import "swagger-ui-react/swagger-ui.css"

const SwaggerUIComponent: React.FC = () => {
  return <SwaggerUI url="http://localhost:8000/api/openapi.yaml" />;
};

export default SwaggerUIComponent;
