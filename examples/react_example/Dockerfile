FROM python:3.9-slim

# Set the working directory
WORKDIR /app

# Copy the current directory contents into the container
COPY . .

# Install required Python packages
RUN pip install --no-cache-dir -r requirements.txt

# Set environment variable for OpenAI API Key (to be provided at runtime)
ENV OPENAI_API_KEY=${OPENAI_API_KEY}

# Use this entrypoint to keep the container running
ENTRYPOINT ["tail", "-f", "/dev/null"]
