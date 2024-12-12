# Sentinel 

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/EntropyLabsAI/sentinel/server)](https://goreportcard.com/report/github.com/EntropyLabsAI/sentinel/server)
[![GitHub stars](https://img.shields.io/github/stars/EntropyLabsAI/sentinel?style=social)](https://github.com/EntropyLabsAI/sentinel/stargazers)
[![PyPI version](https://badge.fury.io/py/entropy-labs.svg)](https://badge.fury.io/py/entropy-labs)
[![Downloads](https://pepy.tech/badge/entropy-labs)](https://pepy.tech/project/entropy-labs)

Sentinel is an agent control plane built by [Entropy Labs](http://entropy-labs.ai/) that allows you to efficiently oversee thousands of agents running in parallel.

🎉 New: [Inspect](https://inspect.ai-safety-institute.org.uk/) has now made approvals a native feature! Check out the Inspect example [here](examples/inspect_example/README.md).

## Sentinel Demo Video
[![Sentinel Demo Video](thumb.png)](https://www.youtube.com/watch?v=pOfnYkdLk18)

🚀 Want to see Sentinel in action or chat about agent supervision? [Book a demo with us]([https://calendly.com/david-mlcoch-entropy-labs/entropy-labs-demo)](https://calendly.com/founders-asteroid-hhaf/30min)!

We're starting with manual reviews for agent actions, but we'll add ways to automatically approve known safe actions in the future.

## Getting Started

See our docs for examples of how to use Sentinel with any agent https://docs.entropy-labs.ai/quickstart

This repo contains a simple web server written in Go and a React frontend. Agent code can make use of our SDK to make requests to our [API](https://docs.entropy-labs.ai/api-reference/project/get-all-projects) when an agent makes tool calls, which will be visible in the Sentinel UI. 

1. Start the webserver and frontend with docker compose:
```bash
cp .env.example .env # Set the environment variables in the .env file
source .env          # Pick up the environment variables
docker compose up    # Start the server and frontend
```

2. Run an agent that is pointing at Sentinel via our [SDK](/entropy_labs/README.md). See the [examples](/examples) for more details.

For more details, see our [docs](https://docs.entropy-labs.ai/introduction).

## Examples
We have a number of example containing agents that are using the Sentinel SDK. These are ready to try out of the box:
- [Langchain](https://docs.entropy-labs.ai/langchain)
- [Inspect](https://docs.entropy-labs.ai/inspect)

## Development

See https://docs.entropy-labs.ai/development
