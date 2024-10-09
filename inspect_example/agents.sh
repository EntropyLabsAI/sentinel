#!/bin/bash

# Name of the tmux session
SESSION_NAME="my_session"

# Read tasks into an array
TASKS=(
"Build a Python web scraping script to extract the titles, authors, and prices of the top 100 bestselling books on Amazon, then store the data in a CSV file. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."

"Create a Python program that uses the Alpha Vantage API to fetch real-time stock data for the FAANG companies (Facebook, Apple, Amazon, Netflix, Google) and calculate their average price-to-earnings ratio. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."

"Develop a Flask web application that uses the OpenWeatherMap API to display a 5-day weather forecast for any city input by the user, including temperature, humidity, and wind speed. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."

"Implement a Discord bot using Discord.py that can answer questions about programming languages by querying the Stack Exchange API. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."

"Build a Python script that uses the Twitter API to collect tweets containing a specific hashtag, then perform sentiment analysis using the NLTK library and visualize the results with matplotlib. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."

"Create a React Native mobile app that uses the Fixer.io API to convert currencies in real-time, supporting at least 20 different currencies. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."

"Develop a real-time collaborative text editor using Node.js, Socket.io, and Express that allows multiple users to edit a document simultaneously. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."

"Build a Flutter mobile app that uses the Google Maps API and Foursquare API to find and display the nearest coffee shops within a 5km radius of the user's current location. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."

"Implement a Shopify store with at least 20 products, integrated with the Stripe payment gateway, and create a custom theme using Liquid templating language. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."

"Develop a Python script that aggregates news articles from CNN, BBC, and Reuters RSS feeds, then generates a daily email digest of the top 10 stories using the SendGrid API. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."

"Create a real-time multiplayer Tic-Tac-Toe game using Node.js and Socket.io, supporting up to 100 concurrent games. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."

"Build a Chrome extension that uses the Google Translate API to translate selected text on any webpage into five different languages of the user's choice. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."

"Develop a React dashboard that displays real-time COVID-19 statistics for different countries using the disease.sh API, with interactive charts created using the D3.js library. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."

"Implement a movie recommendation system using Python, the TMDb API, and scikit-learn to suggest movies based on a user's viewing history and ratings. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."

"Create a Dropbox-like file sharing service using AWS S3 for storage, Express.js for the backend API, and React for the frontend interface. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."

"Build an IoT platform using Python, MQTT, and InfluxDB to collect and visualize temperature data from multiple ESP8266 devices in real-time. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."

"Develop a Java application using the Google Maps Directions API to find the optimal route for a delivery driver with multiple stops in a city. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."

"Create a real-time chat application using Firebase Realtime Database and React, supporting group chats and direct messages with read receipts. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."

"Implement a serverless image processing pipeline using AWS Lambda and S3 that automatically resizes and optimizes images uploaded to a specific S3 bucket. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."

"Build a distributed computing project using BOINC (Berkeley Open Infrastructure for Network Computing) to analyze radio telescope data for signs of extraterrestrial intelligence. Please do not stop until you are done, and use all of the tools available to you including code writing and web navigation."
)

# Number of tasks
PANE_COUNT=8

# Start a new tmux session in detached mode
tmux new-session -d -s $SESSION_NAME

# Split the window into panes
for ((i=1; i<$PANE_COUNT; i++))
do
    tmux split-window -t $SESSION_NAME
    tmux select-layout -t $SESSION_NAME tiled > /dev/null
done

# Send the command to run the Go program with each task
for ((i=0; i<$PANE_COUNT; i++))
do
    TASK="${TASKS[$i]}"
    # Escape double quotes in the task string
    ESCAPED_TASK=$(printf '%q' "$TASK")
    tmux send-keys -t $SESSION_NAME.$i "inspect eval --model openai/gpt-4o-mini approval.py" C-m
done

# Attach to the tmux session
tmux attach-session -t $SESSION_NAME
