# Sentinel 

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/EntropyLabsAI/sentinel/server)](https://goreportcard.com/report/github.com/EntropyLabsAI/sentinel/server)
[![GitHub stars](https://img.shields.io/github/stars/EntropyLabsAI/sentinel?style=social)](https://github.com/EntropyLabsAI/sentinel/stargazers)
[![PyPI version](https://badge.fury.io/py/entropy-labs.svg)](https://badge.fury.io/py/entropy-labs)

Sentinel is an agent control plane built by [Entropy Labs](http://entropy-labs.ai/) that allows you to efficiently oversee thousands of agents running in parallel.

🎉 New: [Inspect](https://inspect.ai-safety-institute.org.uk/) has now made approvals a native feature! Check out the Inspect example [here](examples/inspect_example/README.md).

<div align="center">
<a target="_blank" href="https://www.loom.com/share/c939b9c0da07421b8a3dd665cac26fda"><img width="60%" alt="video thumbnail showing editor" src="./thumb.png"></a>
</div>

🚀 Want to see Sentinel in action or chat about agent supervision? [Book a demo with us](https://calendly.com/david-mlcoch-entropy-labs/entropy-labs-demo)!

We're starting with manual reviews for agent actions, but we'll add ways to automatically approve known safe actions in the future.

## Getting Started

This repo contains a simple web server written in Go and a React frontend. The frontend connects to the server via a websocket and displays reviews that need to be approved. Reviews are submitted to the server via the `/api/review/human` endpoint, and their status is polled from the `/api/review/status` endpoint.

From the root of the repo:

1. Start the webserver and frontend with docker compose:
```bash
cp .env.example .env # Set the environment variables in the .env file
source .env          # Pick up the environment variables
docker compose up    # Start the server and frontend
```

## Examples
Any agent can be run through Sentinel by sending a review to the `/api/review` endpoint and then checking the status of the review with the `/api/review/status` endpoint. Below we show an example of how to do this using curl, and then an example of how to use Sentinel with AISI's Inspect framework.

### [1] Send a review to the interface using curl

1. Send a review to the interface via the `/api/review/human` endpoint:

```bash
curl -X POST http://localhost:8080/api/review/human \
     -H "Content-Type: application/json" \
     -d @examples/curl_example/payload.json
```
2. Check the status of the review programmatically with the `/api/review/status` endpoint:

```bash
curl http://localhost:8080/api/review/status?id=<review-id>
```
(replacing `<review-id>` with the ID of the review you submitted)

3. Navigate to http://localhost:3000 to see the review you submitted and to approve or reject it.


### [2] Run the Inspect example
Inspect is an agent evaluation framework that allows you to evaluate and control agents. We have an example of how to use Inspect with Sentinel [here](examples/inspect_example/README.md).

1. Make sure Inspect AI and Entropy Labs are installed in your python environment:

   ```bash
   pip install inspect-ai entropy_labs --upgrade
   ```

2. Change to the example directory:

   ```bash
   cd examples/inspect_example
   ```

3. Run the example:

   ```bash
   inspect eval run.py --approval approval_human.yaml --model openai/gpt-4o --trace
   ```
This will run the example and trigger the approvals. The example in run.py is choosing random tasks to run from the list of tasks (e.g. build a web app, build a mobile app, etc). It then runs the task and triggers the approval configuration. You should see the approvals in the approval api interface at http://localhost:3000.

There is more information on the Inspect example [here](examples/inspect_example/README.md).

## Development

Make updates to the frontend and backend locally.
