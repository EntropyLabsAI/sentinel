import { defineConfig } from 'orval';

export default defineConfig({
  api: {
    output: {
      client: 'react-query',
      target: './src/types.ts',
    },
    input: '../server/openapi.yaml',
  },
});
